package officialfw

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/blrvio/go-unifi/v11/unifi"
	"github.com/blrvio/go-unifi/v11/unifi/official"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
	Name         types.String `tfsdk:"name"`
	NetworkIDs   types.List   `tfsdk:"network_ids"`
	Origin       types.String `tfsdk:"origin"`
	Configurable types.Bool   `tfsdk:"configurable"`
}

const zoneOriginSystemDefined = "SYSTEM_DEFINED"

// zoneConfigurable reads the metadata "configurable" flag the controller returns for
// system zones (not in the 11.0.99 spec; kept in the metadata union). User zones are
// always configurable; a system zone without the flag reports null.
func zoneConfigurable(meta official.UserOrSystemDefinedEntityMetadata) types.Bool {
	if meta.Origin != zoneOriginSystemDefined {
		return types.BoolValue(true)
	}
	raw, err := json.Marshal(meta)
	if err != nil {
		return types.BoolNull()
	}
	var m struct {
		Configurable *bool `json:"configurable"`
	}
	if json.Unmarshal(raw, &m) != nil || m.Configurable == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*m.Configurable)
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
	m.Origin = types.StringValue(zone.Metadata.Origin)
	m.Configurable = zoneConfigurable(zone.Metadata)

	// The controller returns zone members in its own order; keep the planned/state order for the same set.
	var prior []string
	if ut.IsDefined(m.NetworkIDs) {
		diags.Append(ut.ListElementsAs(ctx, m.NetworkIDs, &prior)...)
	}
	list, listDiags := types.ListValueFrom(ctx, types.StringType, ut.KeepStringOrder(prior, uuidsToStrings(zone.NetworkIds)))
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
			"Requires a controller running version 10.1.78 or later with API-key authentication.\n\n" +
			"The built-in zones (`origin = \"SYSTEM_DEFINED\"`: Internal, External, Gateway, Vpn, Hotspot, Dmz) can be imported. " +
			"Their name cannot change and they cannot be destroyed; only the networks of a zone with `configurable = true` " +
			"can be edited. The plan fails early for any other change; to stop managing a system zone use a `removed` block.",
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
			"origin": schema.StringAttribute{
				MarkdownDescription: "Who defined the zone: `USER_DEFINED` or `SYSTEM_DEFINED`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"configurable": schema.BoolAttribute{
				MarkdownDescription: "Whether the zone's networks can be changed. Always true for user zones; false for the " +
					"External, Gateway and Vpn system zones.",
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *officialFirewallZoneResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	resp.Diagnostics.Append(checkSystemZoneChange(ctx, req)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Destroy plans carry a null plan; nothing else to validate.
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

// checkSystemZoneChange fails the plan for the changes the controller refuses on a
// system zone: destroy, rename, or new networks on a non-configurable zone.
func checkSystemZoneChange(ctx context.Context, req resource.ModifyPlanRequest) diag.Diagnostics {
	var diags diag.Diagnostics
	if req.State.Raw.IsNull() {
		return diags
	}
	var state officialFirewallZoneModel
	diags.Append(req.State.Get(ctx, &state)...)
	if diags.HasError() || state.Origin.ValueString() != zoneOriginSystemDefined {
		return diags
	}
	if req.Plan.Raw.IsNull() {
		diags.AddError("System firewall zone cannot be destroyed",
			fmt.Sprintf("Zone %q is defined by the controller. Stop managing it with a `removed { from = ... lifecycle { destroy = false } }` "+
				"block instead of deleting it.", state.Name.ValueString()))
		return diags
	}
	var plan officialFirewallZoneModel
	diags.Append(req.Plan.Get(ctx, &plan)...)
	if diags.HasError() {
		return diags
	}
	if !plan.Name.IsUnknown() && plan.Name.ValueString() != state.Name.ValueString() {
		diags.AddAttributeError(path.Root("name"), "System firewall zone cannot be renamed",
			fmt.Sprintf("Zone %q is defined by the controller and keeps its name.", state.Name.ValueString()))
	}
	if !state.Configurable.IsNull() && !state.Configurable.ValueBool() && !plan.NetworkIDs.IsUnknown() && !plan.NetworkIDs.Equal(state.NetworkIDs) {
		diags.AddAttributeError(path.Root("network_ids"), "System firewall zone is not configurable",
			fmt.Sprintf("The controller manages the networks of zone %q (configurable = false).", state.Name.ValueString()))
	}
	return diags
}
