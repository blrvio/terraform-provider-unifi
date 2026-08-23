package trafficlist

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// Traffic-matching-list outer type discriminators, matching the go-unifi Official API union.
const (
	tmlTypeIPv4 = "IPV4_ADDRESSES"
	tmlTypeIPv6 = "IPV6_ADDRESSES"
	tmlTypePort = "PORTS"
)

// Item match-type discriminators, matching the inner Matching unions.
const (
	matchIPAddress      = "IP_ADDRESS"
	matchIPAddressRange = "IP_ADDRESS_RANGE"
	matchSubnet         = "SUBNET"
	matchPortNumber     = "PORT_NUMBER"
	matchPortRange      = "PORT_NUMBER_RANGE"
)

var (
	_ resource.Resource                = &trafficMatchingListResource{}
	_ resource.ResourceWithConfigure   = &trafficMatchingListResource{}
	_ resource.ResourceWithImportState = &trafficMatchingListResource{}
	_ resource.ResourceWithModifyPlan  = &trafficMatchingListResource{}
	_ base.ResourceModel               = &trafficMatchingListModel{}
	_ base.Resource                    = &trafficMatchingListResource{}
)

// matchTypesForOuter maps each outer TML type to the inner match types it allows.
var matchTypesForOuter = map[string][]string{
	tmlTypeIPv4: {matchIPAddress, matchIPAddressRange, matchSubnet},
	tmlTypeIPv6: {matchIPAddress, matchSubnet},
	tmlTypePort: {matchPortNumber, matchPortRange},
}

// rangeMatchTypes are the inner match types that use start/stop rather than value.
var rangeMatchTypes = map[string]bool{
	matchIPAddressRange: true,
	matchPortRange:      true,
}

// tmlItemModel is one entry in a traffic matching list. Value/start/stop are all
// modeled as strings uniformly across TML types; PORTS values are parsed to
// int32 when building the API body.
type tmlItemModel struct {
	MatchType types.String `tfsdk:"match_type"`
	Value     types.String `tfsdk:"value"`
	Start     types.String `tfsdk:"start"`
	Stop      types.String `tfsdk:"stop"`
}

func (m tmlItemModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"match_type": types.StringType,
		"value":      types.StringType,
		"start":      types.StringType,
		"stop":       types.StringType,
	}
}

func tmlItemObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: tmlItemModel{}.AttributeTypes()}
}

type trafficMatchingListModel struct {
	base.Model
	Type  types.String `tfsdk:"type"`
	Name  types.String `tfsdk:"name"`
	Items types.List   `tfsdk:"items"`
}

func strPtr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

func strFromPtr(p *string) types.String {
	if p == nil {
		return types.StringNull()
	}
	return types.StringValue(*p)
}

// int32PtrFromString parses a decimal string into an *int32 for PORTS items.
func int32PtrFromString(v types.String) (*int32, error) {
	if v.IsNull() || v.IsUnknown() {
		return nil, nil
	}
	n, err := strconv.Atoi(v.ValueString())
	if err != nil {
		return nil, err
	}
	if n < 1 || n > 65535 {
		return nil, fmt.Errorf("port %d out of range (1-65535)", n)
	}
	i := int32(n) //nolint:gosec // G109: n is range-checked to 1-65535 immediately above
	return &i, nil
}

func strFromInt32Ptr(p *int32) types.String {
	if p == nil {
		return types.StringNull()
	}
	return types.StringValue(strconv.Itoa(int(*p)))
}

// AsUnifiModel builds the Official-API create/update body for the configured TML
// type. Note that TrafficMatchingListCreateOrUpdate.MarshalJSON overlays the
// top-level Name/Type over the union payload, so Name must be set on the body in
// addition to the variant (the From* helpers set only the union + Type).
func (m *trafficMatchingListModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := &official.TrafficMatchingListCreateOrUpdate{}
	name := m.Name.ValueString()

	var items []tmlItemModel
	diags.Append(ut.ListElementsAs(ctx, m.Items, &items)...)
	if diags.HasError() {
		return nil, diags
	}

	var err error
	switch m.Type.ValueString() {
	case tmlTypeIPv4:
		matchings := make([]official.IPv4Matching, 0, len(items))
		for _, it := range items {
			var mm official.IPv4Matching
			switch it.MatchType.ValueString() {
			case matchIPAddress:
				err = mm.FromAddressIPv4Matching(official.AddressIPv4Matching{Value: strPtr(it.Value)})
			case matchIPAddressRange:
				err = mm.FromAddressRangeIPv4Matching(official.AddressRangeIPv4Matching{Start: strPtr(it.Start), Stop: strPtr(it.Stop)})
			case matchSubnet:
				err = mm.FromSubnetIPv4Matching(official.SubnetIPv4Matching{Value: strPtr(it.Value)})
			default:
				diags.AddError("Unsupported match_type", fmt.Sprintf("match_type %q is not valid for type %q", it.MatchType.ValueString(), tmlTypeIPv4))
				return nil, diags
			}
			if err != nil {
				diags.AddError("Failed to build IPv4 matching item", err.Error())
				return nil, diags
			}
			matchings = append(matchings, mm)
		}
		err = body.FromIpV4TrafficMatchingListCreateUpdate(official.IpV4TrafficMatchingListCreateUpdate{
			Name:  name,
			Items: &matchings,
		})
	case tmlTypeIPv6:
		matchings := make([]official.IPv6Matching, 0, len(items))
		for _, it := range items {
			var mm official.IPv6Matching
			switch it.MatchType.ValueString() {
			case matchIPAddress:
				err = mm.FromAddressIPv6Matching(official.AddressIPv6Matching{Value: strPtr(it.Value)})
			case matchSubnet:
				err = mm.FromSubnetIPv6Matching(official.SubnetIPv6Matching{Value: strPtr(it.Value)})
			default:
				diags.AddError("Unsupported match_type", fmt.Sprintf("match_type %q is not valid for type %q", it.MatchType.ValueString(), tmlTypeIPv6))
				return nil, diags
			}
			if err != nil {
				diags.AddError("Failed to build IPv6 matching item", err.Error())
				return nil, diags
			}
			matchings = append(matchings, mm)
		}
		err = body.FromIpV6TrafficMatchingListCreateUpdate(official.IpV6TrafficMatchingListCreateUpdate{
			Name:  name,
			Items: &matchings,
		})
	case tmlTypePort:
		matchings := make([]official.PortMatching, 0, len(items))
		for _, it := range items {
			var mm official.PortMatching
			switch it.MatchType.ValueString() {
			case matchPortNumber:
				value, convErr := int32PtrFromString(it.Value)
				if convErr != nil {
					diags.AddError("Invalid port value", fmt.Sprintf("value %q must be an integer", it.Value.ValueString()))
					return nil, diags
				}
				err = mm.FromNumberPortMatching(official.NumberPortMatching{Value: value})
			case matchPortRange:
				start, startErr := int32PtrFromString(it.Start)
				stop, stopErr := int32PtrFromString(it.Stop)
				if startErr != nil || stopErr != nil {
					diags.AddError("Invalid port range", "start and stop must be integers")
					return nil, diags
				}
				err = mm.FromNumberRangePortMatching(official.NumberRangePortMatching{Start: start, Stop: stop})
			default:
				diags.AddError("Unsupported match_type", fmt.Sprintf("match_type %q is not valid for type %q", it.MatchType.ValueString(), tmlTypePort))
				return nil, diags
			}
			if err != nil {
				diags.AddError("Failed to build port matching item", err.Error())
				return nil, diags
			}
			matchings = append(matchings, mm)
		}
		err = body.FromPortTrafficMatchingListCreateUpdate(official.PortTrafficMatchingListCreateUpdate{
			Name:  name,
			Items: &matchings,
		})
	default:
		diags.AddError("Unsupported traffic matching list type", fmt.Sprintf("Unknown type %q", m.Type.ValueString()))
		return nil, diags
	}
	if err != nil {
		diags.AddError("Failed to build traffic matching list request", err.Error())
		return nil, diags
	}
	// The union From* helpers set only the union payload + Type; MarshalJSON
	// overlays the top-level Name, so it must be set explicitly or every request
	// would send an empty name.
	body.Name = name
	return body, diags
}

// asTMLBody performs the checked type assertion from the AsUnifiModel interface{}
// return to the concrete Official-API body.
func asTMLBody(body interface{}) (*official.TrafficMatchingListCreateOrUpdate, diag.Diagnostics) {
	var diags diag.Diagnostics
	b, ok := body.(*official.TrafficMatchingListCreateOrUpdate)
	if !ok {
		diags.AddError("Internal error", fmt.Sprintf("expected *official.TrafficMatchingListCreateOrUpdate, got %T", body))
	}
	return b, diags
}

// Merge populates the model from an Official-API TrafficMatchingList read response.
func (m *trafficMatchingListModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tml, ok := other.(*official.TrafficMatchingList)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *official.TrafficMatchingList, got %T", other))
		return diags
	}

	m.SetID(tml.Id.String())
	m.Type = types.StringValue(tml.Type)
	m.Name = types.StringValue(tml.Name)

	items := make([]tmlItemModel, 0)
	switch tml.Type {
	case tmlTypeIPv4:
		v, err := tml.AsIpV4TrafficMatchingList()
		if err != nil {
			diags.AddError("Failed to decode IPv4 traffic matching list", err.Error())
			return diags
		}
		if v.Items != nil {
			for _, it := range *v.Items {
				item := tmlItemModel{MatchType: types.StringValue(it.Type)}
				switch it.Type {
				case matchIPAddress:
					iv, decErr := it.AsAddressIPv4Matching()
					if decErr != nil {
						diags.AddError("Failed to decode IPv4 address item", decErr.Error())
						return diags
					}
					item.Value = strFromPtr(iv.Value)
				case matchIPAddressRange:
					iv, decErr := it.AsAddressRangeIPv4Matching()
					if decErr != nil {
						diags.AddError("Failed to decode IPv4 address range item", decErr.Error())
						return diags
					}
					item.Start = strFromPtr(iv.Start)
					item.Stop = strFromPtr(iv.Stop)
				case matchSubnet:
					iv, decErr := it.AsSubnetIPv4Matching()
					if decErr != nil {
						diags.AddError("Failed to decode IPv4 subnet item", decErr.Error())
						return diags
					}
					item.Value = strFromPtr(iv.Value)
				default:
					diags.AddError("Unsupported match_type", fmt.Sprintf("Controller returned unknown IPv4 match type %q", it.Type))
					return diags
				}
				items = append(items, item)
			}
		}
	case tmlTypeIPv6:
		v, err := tml.AsIpV6TrafficMatchingList()
		if err != nil {
			diags.AddError("Failed to decode IPv6 traffic matching list", err.Error())
			return diags
		}
		if v.Items != nil {
			for _, it := range *v.Items {
				item := tmlItemModel{MatchType: types.StringValue(it.Type)}
				switch it.Type {
				case matchIPAddress:
					iv, decErr := it.AsAddressIPv6Matching()
					if decErr != nil {
						diags.AddError("Failed to decode IPv6 address item", decErr.Error())
						return diags
					}
					item.Value = strFromPtr(iv.Value)
				case matchSubnet:
					iv, decErr := it.AsSubnetIPv6Matching()
					if decErr != nil {
						diags.AddError("Failed to decode IPv6 subnet item", decErr.Error())
						return diags
					}
					item.Value = strFromPtr(iv.Value)
				default:
					diags.AddError("Unsupported match_type", fmt.Sprintf("Controller returned unknown IPv6 match type %q", it.Type))
					return diags
				}
				items = append(items, item)
			}
		}
	case tmlTypePort:
		v, err := tml.AsPortTrafficMatchingList()
		if err != nil {
			diags.AddError("Failed to decode port traffic matching list", err.Error())
			return diags
		}
		if v.Items != nil {
			for _, it := range *v.Items {
				item := tmlItemModel{MatchType: types.StringValue(it.Type)}
				switch it.Type {
				case matchPortNumber:
					iv, decErr := it.AsNumberPortMatching()
					if decErr != nil {
						diags.AddError("Failed to decode port number item", decErr.Error())
						return diags
					}
					item.Value = strFromInt32Ptr(iv.Value)
				case matchPortRange:
					iv, decErr := it.AsNumberRangePortMatching()
					if decErr != nil {
						diags.AddError("Failed to decode port range item", decErr.Error())
						return diags
					}
					item.Start = strFromInt32Ptr(iv.Start)
					item.Stop = strFromInt32Ptr(iv.Stop)
				default:
					diags.AddError("Unsupported match_type", fmt.Sprintf("Controller returned unknown port match type %q", it.Type))
					return diags
				}
				items = append(items, item)
			}
		}
	default:
		diags.AddError("Unsupported traffic matching list type", fmt.Sprintf("Controller returned unknown type %q", tml.Type))
		return diags
	}

	list, listDiags := types.ListValueFrom(ctx, tmlItemObjectType(), items)
	diags.Append(listDiags...)
	m.Items = list
	return diags
}

type trafficMatchingListResource struct {
	*base.GenericResource[*trafficMatchingListModel]
}

// NewTrafficMatchingListResource creates the unifi_traffic_matching_list resource.
// It embeds a GenericResource purely for Configure/version-validator wiring; CRUD
// is custom because the Official API is keyed by site/entity UUID.
func NewTrafficMatchingListResource() resource.Resource {
	r := &trafficMatchingListResource{}
	r.GenericResource = base.NewGenericResource(
		"unifi_traffic_matching_list",
		func() *trafficMatchingListModel { return &trafficMatchingListModel{} },
		base.ResourceFunctions{},
	)
	return r
}

func (r *trafficMatchingListResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a traffic matching list (a reusable set of IPv4 addresses, IPv6 addresses " +
			"or ports) via the UniFi **Official API** (`integration/v1`). Requires a controller running " +
			"version 10.1.78 or later with API-key authentication.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"type": schema.StringAttribute{
				MarkdownDescription: "The traffic matching list type. One of `IPV4_ADDRESSES`, `IPV6_ADDRESSES`, " +
					"`PORTS`. Changing the type forces a new resource.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(tmlTypeIPv4, tmlTypeIPv6, tmlTypePort),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the traffic matching list.",
				Required:            true,
			},
			"items": schema.ListNestedAttribute{
				MarkdownDescription: "The entries that make up the matching list.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"match_type": schema.StringAttribute{
							MarkdownDescription: "The item match type. For `IPV4_ADDRESSES`: `IP_ADDRESS`, " +
								"`IP_ADDRESS_RANGE` or `SUBNET`. For `IPV6_ADDRESSES`: `IP_ADDRESS` or `SUBNET`. " +
								"For `PORTS`: `PORT_NUMBER` or `PORT_NUMBER_RANGE`.",
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf(
									matchIPAddress, matchIPAddressRange, matchSubnet,
									matchPortNumber, matchPortRange,
								),
							},
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "The single value for non-range match types (`IP_ADDRESS`, " +
								"`SUBNET`, `PORT_NUMBER`). For `PORT_NUMBER`, provide the port as a string.",
							Optional: true,
						},
						"start": schema.StringAttribute{
							MarkdownDescription: "The start value for range match types (`IP_ADDRESS_RANGE`, " +
								"`PORT_NUMBER_RANGE`). For `PORT_NUMBER_RANGE`, provide the port as a string.",
							Optional: true,
						},
						"stop": schema.StringAttribute{
							MarkdownDescription: "The stop value for range match types (`IP_ADDRESS_RANGE`, " +
								"`PORT_NUMBER_RANGE`). For `PORT_NUMBER_RANGE`, provide the port as a string.",
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func (r *trafficMatchingListResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Destroy plans carry a null plan; nothing to validate.
	if req.Plan.Raw.IsNull() {
		return
	}
	resp.Diagnostics.Append(r.GetClient().RequireOfficialAPI()...)
	resp.Diagnostics.Append(r.RequireMinVersion(base.ControllerVersionOfficialAPI.String())...)

	var plan trafficMatchingListModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	outerType := plan.Type.ValueString()
	if outerType == "" {
		return
	}
	valid := matchTypesForOuter[outerType]

	var items []tmlItemModel
	resp.Diagnostics.Append(ut.ListElementsAs(ctx, plan.Items, &items)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for i, it := range items {
		matchType := it.MatchType.ValueString()
		if matchType == "" {
			continue
		}
		if !contains(valid, matchType) {
			resp.Diagnostics.AddAttributeError(
				path.Root("items").AtListIndex(i).AtName("match_type"),
				"Invalid match_type",
				fmt.Sprintf("match_type %q is not valid for type %q. Valid values are %v.", matchType, outerType, valid),
			)
			continue
		}
		if rangeMatchTypes[matchType] {
			if ut.IsEmptyString(it.Start) || ut.IsEmptyString(it.Stop) {
				resp.Diagnostics.AddAttributeError(
					path.Root("items").AtListIndex(i),
					"Missing range bounds",
					fmt.Sprintf("`start` and `stop` are required when `match_type` is %q.", matchType),
				)
			}
		} else if ut.IsEmptyString(it.Value) {
			resp.Diagnostics.AddAttributeError(
				path.Root("items").AtListIndex(i),
				"Missing value",
				fmt.Sprintf("`value` is required when `match_type` is %q.", matchType),
			)
		}
	}
}

func contains(s []string, v string) bool {
	for _, e := range s {
		if e == v {
			return true
		}
	}
	return false
}

func (r *trafficMatchingListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan trafficMatchingListModel
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
	createBody, diags := asTMLBody(body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := client.Official().TrafficMatchingLists().Create(ctx, siteID, *createBody)
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error creating traffic matching list", err)...)
		return
	}
	resp.Diagnostics.Append(plan.Merge(ctx, created)...)
	plan.SetSite(siteName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *trafficMatchingListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state trafficMatchingListModel
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
	tml, err := client.Official().TrafficMatchingLists().Get(ctx, siteID, id)
	if err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error reading traffic matching list", err)...)
		return
	}
	resp.Diagnostics.Append(state.Merge(ctx, tml)...)
	state.SetSite(siteName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *trafficMatchingListResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan, state trafficMatchingListModel
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
	updateBody, diags := asTMLBody(body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := client.Official().TrafficMatchingLists().Update(ctx, siteID, id, *updateBody)
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error updating traffic matching list", err)...)
		return
	}
	resp.Diagnostics.Append(plan.Merge(ctx, updated)...)
	plan.SetSite(siteName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *trafficMatchingListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state trafficMatchingListModel
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
	if err := client.Official().TrafficMatchingLists().Delete(ctx, siteID, id); err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			return
		}
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error deleting traffic matching list", err)...)
	}
}

func (r *trafficMatchingListResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, site := base.ImportIDWithSite(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site"), site)...)
}
