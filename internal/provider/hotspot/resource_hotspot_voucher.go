package hotspot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

var (
	_ resource.Resource                = &hotspotVoucherResource{}
	_ resource.ResourceWithConfigure   = &hotspotVoucherResource{}
	_ resource.ResourceWithImportState = &hotspotVoucherResource{}
	_ resource.ResourceWithModifyPlan  = &hotspotVoucherResource{}
	_ base.ResourceModel               = &hotspotVoucherModel{}
	_ base.Resource                    = &hotspotVoucherResource{}
)

type hotspotVoucherModel struct {
	base.Model
	// Configurable (all force-replace: the Official API cannot update a voucher).
	Name                 types.String `tfsdk:"name"`
	TimeLimitMinutes     types.Int64  `tfsdk:"time_limit_minutes"`
	AuthorizedGuestLimit types.Int64  `tfsdk:"authorized_guest_limit"`
	DataUsageLimitMBytes types.Int64  `tfsdk:"data_usage_limit_mbytes"`
	RxRateLimitKbps      types.Int64  `tfsdk:"rx_rate_limit_kbps"`
	TxRateLimitKbps      types.Int64  `tfsdk:"tx_rate_limit_kbps"`
	// Computed (populated from GetVoucher after create).
	Code                 types.String `tfsdk:"code"`
	CreatedAt            types.String `tfsdk:"created_at"`
	ActivatedAt          types.String `tfsdk:"activated_at"`
	ExpiresAt            types.String `tfsdk:"expires_at"`
	Expired              types.Bool   `tfsdk:"expired"`
	AuthorizedGuestCount types.Int64  `tfsdk:"authorized_guest_count"`
}

func int64Ptr(v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := v.ValueInt64()
	return &i
}

func int64FromPtr(p *int64) types.Int64 {
	if p == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*p)
}

func timeToString(t time.Time) types.String {
	if t.IsZero() {
		return types.StringNull()
	}
	return types.StringValue(t.Format(time.RFC3339))
}

func timePtrToString(p *time.Time) types.String {
	if p == nil {
		return types.StringNull()
	}
	return timeToString(*p)
}

// AsUnifiModel builds the Official-API voucher creation request. Count is always
// 1: this resource manages exactly one generated voucher.
func (m *hotspotVoucherModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	count := int32(1)
	body := &official.HotspotVoucherCreationRequest{
		Count:                &count,
		Name:                 m.Name.ValueString(),
		TimeLimitMinutes:     m.TimeLimitMinutes.ValueInt64(),
		AuthorizedGuestLimit: int64Ptr(m.AuthorizedGuestLimit),
		DataUsageLimitMBytes: int64Ptr(m.DataUsageLimitMBytes),
		RxRateLimitKbps:      int64Ptr(m.RxRateLimitKbps),
		TxRateLimitKbps:      int64Ptr(m.TxRateLimitKbps),
	}
	return body, nil
}

// asVoucherCreationRequest performs the checked type assertion from the
// AsUnifiModel interface{} return to the concrete Official-API body.
func asVoucherCreationRequest(body interface{}) (*official.HotspotVoucherCreationRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	b, ok := body.(*official.HotspotVoucherCreationRequest)
	if !ok {
		diags.AddError("Internal error", fmt.Sprintf("expected *official.HotspotVoucherCreationRequest, got %T", body))
	}
	return b, diags
}

// Merge populates the model from an Official-API voucher read response.
func (m *hotspotVoucherModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	v, ok := other.(*official.HotspotVoucherDetails)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *official.HotspotVoucherDetails, got %T", other))
		return diags
	}

	m.SetID(v.Id.String())
	m.Name = types.StringValue(v.Name)
	m.TimeLimitMinutes = types.Int64Value(v.TimeLimitMinutes)
	m.AuthorizedGuestLimit = int64FromPtr(v.AuthorizedGuestLimit)
	m.DataUsageLimitMBytes = int64FromPtr(v.DataUsageLimitMBytes)
	m.RxRateLimitKbps = int64FromPtr(v.RxRateLimitKbps)
	m.TxRateLimitKbps = int64FromPtr(v.TxRateLimitKbps)

	m.Code = types.StringValue(v.Code)
	m.CreatedAt = timeToString(v.CreatedAt)
	m.ActivatedAt = timePtrToString(v.ActivatedAt)
	m.ExpiresAt = timePtrToString(v.ExpiresAt)
	m.Expired = types.BoolValue(v.Expired)
	m.AuthorizedGuestCount = types.Int64Value(v.AuthorizedGuestCount)
	return diags
}

type hotspotVoucherResource struct {
	*base.GenericResource[*hotspotVoucherModel]
}

// NewHotspotVoucherResource creates the unifi_hotspot_voucher resource. It embeds
// a GenericResource purely for Configure/version-validator wiring; CRUD is custom
// because the Official API is keyed by site/entity UUID and vouchers are
// create-only (no update).
func NewHotspotVoucherResource() resource.Resource {
	r := &hotspotVoucherResource{}
	r.GenericResource = base.NewGenericResource(
		"unifi_hotspot_voucher",
		func() *hotspotVoucherModel { return &hotspotVoucherModel{} },
		base.ResourceFunctions{},
	)
	return r
}

func (r *hotspotVoucherResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Hotspot guest voucher via the UniFi **Official API** (`integration/v1`). " +
			"Requires a controller running version 10.1.78 or later with API-key authentication. " +
			"Vouchers are **create-only**: the Official API has no update operation, so changing any argument " +
			"forces the voucher to be destroyed and recreated (which generates a new code).",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "Voucher note. Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"time_limit_minutes": schema.Int64Attribute{
				MarkdownDescription: "How long (in minutes) the voucher provides access after the first guest " +
					"authorizes. Changing this forces a new resource.",
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{int64validator.Between(1, 1000000)},
			},
			"authorized_guest_limit": schema.Int64Attribute{
				MarkdownDescription: "Limit for how many different guests can use the same voucher to authorize " +
					"network access. Changing this forces a new resource.",
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{int64validator.AtLeast(1)},
			},
			"data_usage_limit_mbytes": schema.Int64Attribute{
				MarkdownDescription: "Data usage limit in megabytes. Changing this forces a new resource.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{int64validator.Between(1, 1048576)},
			},
			"rx_rate_limit_kbps": schema.Int64Attribute{
				MarkdownDescription: "Download rate limit in kilobits per second. Changing this forces a new resource.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{int64validator.Between(2, 100000)},
			},
			"tx_rate_limit_kbps": schema.Int64Attribute{
				MarkdownDescription: "Upload rate limit in kilobits per second. Changing this forces a new resource.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{int64validator.Between(2, 100000)},
			},
			"code": schema.StringAttribute{
				MarkdownDescription: "The secret code used to activate the voucher via the Hotspot portal.",
				Computed:            true,
				Sensitive:           true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp (RFC3339) when the voucher was created.",
				Computed:            true,
			},
			"activated_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp (RFC3339) when the voucher was activated (first guest authorized), " +
					"or null if not yet activated.",
				Computed: true,
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp (RFC3339) when the voucher expires, or null if not yet determined.",
				Computed:            true,
			},
			"expired": schema.BoolAttribute{
				MarkdownDescription: "Whether the voucher has expired and can no longer authorize network access.",
				Computed:            true,
			},
			"authorized_guest_count": schema.Int64Attribute{
				MarkdownDescription: "How many guests have used the voucher to authorize network access.",
				Computed:            true,
			},
		},
	}
}

func (r *hotspotVoucherResource) ModifyPlan(_ context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Destroy plans carry a null plan; nothing to validate.
	if req.Plan.Raw.IsNull() {
		return
	}
	resp.Diagnostics.Append(r.GetClient().RequireOfficialAPI()...)
	resp.Diagnostics.Append(r.RequireMinVersion(base.ControllerVersionOfficialAPI.String())...)
}

func (r *hotspotVoucherResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan hotspotVoucherModel
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
	createBody, diags := asVoucherCreationRequest(body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := client.Official().Hotspot().CreateVouchers(ctx, siteID, *createBody)
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error creating hotspot voucher", err)...)
		return
	}
	voucherID, diags := createdVoucherID(result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Re-read the voucher to populate every computed field consistently.
	voucher, err := client.Official().Hotspot().GetVoucher(ctx, siteID, voucherID)
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error reading created hotspot voucher", err)...)
		return
	}
	resp.Diagnostics.Append(plan.Merge(ctx, voucher)...)
	plan.SetSite(siteName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// createdVoucherID extracts the single generated voucher's id from the creation
// result (Count is always 1).
func createdVoucherID(result *official.VoucherCreationResult) (uuid.UUID, diag.Diagnostics) {
	var diags diag.Diagnostics
	if result == nil || result.Vouchers == nil || len(*result.Vouchers) == 0 {
		diags.AddError(
			"Voucher creation returned no voucher",
			"The UniFi Official API returned no generated voucher in the creation result.",
		)
		return uuid.UUID{}, diags
	}
	return (*result.Vouchers)[0].Id, diags
}

func (r *hotspotVoucherResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state hotspotVoucherModel
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
	voucher, err := client.Official().Hotspot().GetVoucher(ctx, siteID, id)
	if err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error reading hotspot voucher", err)...)
		return
	}
	resp.Diagnostics.Append(state.Merge(ctx, voucher)...)
	state.SetSite(siteName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is a no-op. Every configurable attribute is RequiresReplace, so a real
// change forces recreation and this handler is never called with mutated values;
// it only re-sets state from the plan to satisfy the resource.Resource interface.
func (r *hotspotVoucherResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan hotspotVoucherModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *hotspotVoucherResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state hotspotVoucherModel
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
	if _, err := client.Official().Hotspot().DeleteVoucher(ctx, siteID, id); err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			return
		}
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error deleting hotspot voucher", err)...)
	}
}

func (r *hotspotVoucherResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, site := base.ImportIDWithSite(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site"), site)...)
}
