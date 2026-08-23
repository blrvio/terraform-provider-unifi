package dashboard

import (
	"context"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// dashboardModuleModel is one module placed on a dashboard.
type dashboardModuleModel struct {
	ID           types.String `tfsdk:"id"`
	ModuleID     types.String `tfsdk:"module_id"`
	Config       types.String `tfsdk:"config"`
	Restrictions types.String `tfsdk:"restrictions"`
}

func (m dashboardModuleModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":           types.StringType,
		"module_id":    types.StringType,
		"config":       types.StringType,
		"restrictions": types.StringType,
	}
}

func dashboardModuleObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: dashboardModuleModel{}.AttributeTypes()}
}

// DashboardModel represents the data model for a UniFi dashboard.
type DashboardModel struct {
	base.Model
	Name              types.String `tfsdk:"name"`
	Desc              types.String `tfsdk:"desc"`
	ControllerVersion types.String `tfsdk:"controller_version"`
	IsPublic          types.Bool   `tfsdk:"is_public"`
	Modules           types.List   `tfsdk:"modules"`
}

func dashboardModulesAsUnifi(ctx context.Context, list types.List) ([]unifi.DashboardModules, diag.Diagnostics) {
	var diags diag.Diagnostics
	var modules []dashboardModuleModel
	diags.Append(ut.ListElementsAs(ctx, list, &modules)...)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]unifi.DashboardModules, 0, len(modules))
	for _, m := range modules {
		out = append(out, unifi.DashboardModules{
			ID:           m.ID.ValueString(),
			ModuleID:     m.ModuleID.ValueString(),
			Config:       m.Config.ValueString(),
			Restrictions: m.Restrictions.ValueString(),
		})
	}
	return out, diags
}

func dashboardModulesToList(ctx context.Context, modules []unifi.DashboardModules) (types.List, diag.Diagnostics) {
	items := make([]dashboardModuleModel, 0, len(modules))
	for _, m := range modules {
		items = append(items, dashboardModuleModel{
			ID:           ut.StringOrNull(m.ID),
			ModuleID:     ut.StringOrNull(m.ModuleID),
			Config:       ut.StringOrNull(m.Config),
			Restrictions: ut.StringOrNull(m.Restrictions),
		})
	}
	return types.ListValueFrom(ctx, dashboardModuleObjectType(), items)
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *DashboardModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	modules, diags := dashboardModulesAsUnifi(ctx, m.Modules)
	if diags.HasError() {
		return nil, diags
	}
	return &unifi.Dashboard{
		ID:                m.ID.ValueString(),
		Name:              m.Name.ValueString(),
		Desc:              m.Desc.ValueString(),
		ControllerVersion: m.ControllerVersion.ValueString(),
		IsPublic:          m.IsPublic.ValueBool(),
		Modules:           modules,
	}, diags
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *DashboardModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	model, ok := other.(*unifi.Dashboard)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.Dashboard, got %T", other))
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.Name = ut.StringOrNull(model.Name)
	m.Desc = ut.StringOrNull(model.Desc)
	m.ControllerVersion = ut.StringOrNull(model.ControllerVersion)
	m.IsPublic = types.BoolValue(model.IsPublic)

	list, listDiags := dashboardModulesToList(ctx, model.Modules)
	diags.Append(listDiags...)
	m.Modules = list
	return diags
}

var (
	_ resource.Resource                = &dashboardResource{}
	_ resource.ResourceWithConfigure   = &dashboardResource{}
	_ resource.ResourceWithImportState = &dashboardResource{}
	_ base.Resource                    = &dashboardResource{}
	_ base.ResourceModel               = &DashboardModel{}
)

type dashboardResource struct {
	*base.GenericResource[*DashboardModel]
}

// NewDashboardResource creates a new instance of the dashboard resource.
func NewDashboardResource() resource.Resource {
	return &dashboardResource{
		GenericResource: base.NewGenericResource(
			"unifi_dashboard",
			func() *DashboardModel { return &DashboardModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetDashboard(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					dashboard, _ := model.(*unifi.Dashboard)
					return client.CreateDashboard(ctx, site, dashboard)
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					dashboard, _ := model.(*unifi.Dashboard)
					return client.UpdateDashboard(ctx, site, dashboard)
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteDashboard(ctx, site, id)
				},
			},
		),
	}
}

func dashboardModuleAttributes() map[string]schema.Attribute {
	strAttr := func(desc string) schema.Attribute {
		return schema.StringAttribute{MarkdownDescription: desc, Optional: true, Computed: true}
	}
	return map[string]schema.Attribute{
		"id":           strAttr("The identifier of the module instance on the dashboard."),
		"module_id":    strAttr("The identifier of the module type."),
		"config":       strAttr("The module configuration payload (opaque JSON string)."),
		"restrictions": strAttr("The module restrictions payload (opaque JSON string)."),
	}
}

// Schema defines the schema for the resource.
func (r *dashboardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_dashboard` resource manages a custom dashboard in the UniFi controller: " +
			"a named, optionally public collection of modules. This is the internal-SDK dashboard entity and is " +
			"distinct from `unifi_setting_dashboard`, which controls the built-in dashboard layout setting.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the dashboard.",
				Optional:            true,
				Computed:            true,
			},
			"desc": schema.StringAttribute{
				MarkdownDescription: "A free-form description of the dashboard.",
				Optional:            true,
				Computed:            true,
			},
			"controller_version": schema.StringAttribute{
				MarkdownDescription: "The controller version the dashboard was authored against.",
				Optional:            true,
				Computed:            true,
			},
			"is_public": schema.BoolAttribute{
				MarkdownDescription: "Whether the dashboard is publicly visible.",
				Optional:            true,
				Computed:            true,
			},
			"modules": schema.ListNestedAttribute{
				MarkdownDescription: "The modules that make up the dashboard.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: dashboardModuleAttributes(),
				},
			},
		},
	}
}
