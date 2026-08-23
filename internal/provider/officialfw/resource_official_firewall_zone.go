package officialfw

import (
	"context"
	"errors"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

var (
	_ resource.Resource                = &officialFirewallZoneResource{}
	_ resource.ResourceWithConfigure   = &officialFirewallZoneResource{}
	_ resource.ResourceWithImportState = &officialFirewallZoneResource{}
	_ resource.ResourceWithModifyPlan  = &officialFirewallZoneResource{}
	_ base.ResourceModel               = &officialFirewallZoneModel{}
	_ base.Resource                    = &officialFirewallZoneResource{}
)

type officialFirewallZoneModel struct {
	base.Model
	Name       types.String `tfsdk:"name"`
	NetworkIDs types.List   `tfsdk:"network_ids"`
}

// stringsToUUIDs parses a slice of string UUIDs into []uuid.UUID, reporting the
// first parse failure as a diagnostic.
func stringsToUUIDs(in []string) ([]uuid.UUID, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := make([]uuid.UUID, 0, len(in))
	for _, s := range in {
		id, err := uuid.Parse(s)
		if err != nil {
			diags.AddError("Invalid UUID", fmt.Sprintf("Could not parse %q as a UUID: %s", s, err))
			return nil, diags
		}
		out = append(out, id)
	}
	return out, diags
}

// uuidsToStrings renders a slice of UUIDs as their canonical string form.
func uuidsToStrings(in []uuid.UUID) []string {
	out := make([]string, 0, len(in))
	for _, id := range in {
		out = append(out, id.String())
	}
	return out
}

// AsUnifiModel builds the Official-API create/update body for a firewall zone.
func (m *officialFirewallZoneModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	var networkStrings []string
	diags.Append(ut.ListElementsAs(ctx, m.NetworkIDs, &networkStrings)...)
	if diags.HasError() {
		return nil, diags
	}
	networkIDs, convDiags := stringsToUUIDs(networkStrings)
	diags.Append(convDiags...)
	if diags.HasError() {
		return nil, diags
	}

	return &official.FirewallZoneCreateOrUpdate{
		Name:       m.Name.ValueString(),
		NetworkIds: networkIDs,
	}, diags
}

// asFirewallZoneBody performs the checked type assertion from the AsUnifiModel
// interface{} return to the concrete Official-API body.
func asFirewallZoneBody(body interface{}) (*official.FirewallZoneCreateOrUpdate, diag.Diagnostics) {
	var diags diag.Diagnostics
	b, ok := body.(*official.FirewallZoneCreateOrUpdate)
	if !ok {
		diags.AddError("Internal error", fmt.Sprintf("expected *official.FirewallZoneCreateOrUpdate, got %T", body))
	}
	return b, diags
}

// Merge populates the model from an Official-API FirewallZone read response.
func (m *officialFirewallZoneModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	zone, ok := other.(*official.FirewallZone)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *official.FirewallZone, got %T", other))
		return diags
	}

	m.SetID(zone.Id.String())
	m.Name = types.StringValue(zone.Name)

	list, listDiags := types.ListValueFrom(ctx, types.StringType, uuidsToStrings(zone.NetworkIds))
	diags.Append(listDiags...)
	if diags.HasError() {
		return diags
	}
	m.NetworkIDs = list
	return diags
}

type officialFirewallZoneResource struct {
	*base.GenericResource[*officialFirewallZoneModel]
}

// NewOfficialFirewallZoneResource creates the unifi_official_firewall_zone
// resource. CRUD is custom because the Official API is keyed by site/zone UUID.
func NewOfficialFirewallZoneResource() resource.Resource {
	r := &officialFirewallZoneResource{}
	r.GenericResource = base.NewGenericResource(
		"unifi_official_firewall_zone",
		func() *officialFirewallZoneModel { return &officialFirewallZoneModel{} },
		base.ResourceFunctions{},
	)
	return r
}

func (r *officialFirewallZoneResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a firewall zone via the UniFi **Official API** (`integration/v1`). " +
			"Requires a controller running version 10.1.78 or later with API-key authentication.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the firewall zone.",
				Required:            true,
			},
			"network_ids": schema.ListAttribute{
				MarkdownDescription: "List of network UUIDs assigned to this firewall zone.",
				Required:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *officialFirewallZoneResource) ModifyPlan(_ context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Destroy plans carry a null plan; nothing to validate.
	if req.Plan.Raw.IsNull() {
		return
	}
	resp.Diagnostics.Append(r.GetClient().RequireOfficialAPI()...)
	resp.Diagnostics.Append(r.RequireMinVersion(base.ControllerVersionOfficialAPI.String())...)
}

func (r *officialFirewallZoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan officialFirewallZoneModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	siteName := client.ResolveSite(&plan)
	siteID, diags := client.ResolveSiteUUID(ctx, siteName)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, diags := plan.AsUnifiModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	createBody, diags := asFirewallZoneBody(body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := client.Official().Firewall().CreateZone(ctx, siteID, *createBody)
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error creating firewall zone", err)...)
		return
	}
	resp.Diagnostics.Append(plan.Merge(ctx, created)...)
	plan.SetSite(siteName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *officialFirewallZoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state officialFirewallZoneModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	siteName := client.ResolveSite(&state)
	siteID, id, diags := client.ResolveSiteAndID(ctx, siteName, state.GetID())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	zone, err := client.Official().Firewall().GetZone(ctx, siteID, id)
	if err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error reading firewall zone", err)...)
		return
	}
	resp.Diagnostics.Append(state.Merge(ctx, zone)...)
	state.SetSite(siteName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *officialFirewallZoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan, state officialFirewallZoneModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	siteName := client.ResolveSite(&plan)
	siteID, id, diags := client.ResolveSiteAndID(ctx, siteName, state.GetID())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, diags := plan.AsUnifiModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateBody, diags := asFirewallZoneBody(body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := client.Official().Firewall().UpdateZone(ctx, siteID, id, *updateBody)
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error updating firewall zone", err)...)
		return
	}
	resp.Diagnostics.Append(plan.Merge(ctx, updated)...)
	plan.SetSite(siteName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *officialFirewallZoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state officialFirewallZoneModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	siteName := client.ResolveSite(&state)
	siteID, id, diags := client.ResolveSiteAndID(ctx, siteName, state.GetID())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := client.Official().Firewall().DeleteZone(ctx, siteID, id); err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			return
		}
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error deleting firewall zone", err)...)
	}
}

func (r *officialFirewallZoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, site := base.ImportIDWithSite(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site"), site)...)
}
