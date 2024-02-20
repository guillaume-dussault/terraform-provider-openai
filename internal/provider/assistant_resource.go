package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	openai "github.com/sashabaranov/go-openai"
	"golang.org/x/exp/slices"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &assistantResource{}
	_ resource.ResourceWithConfigure   = &assistantResource{}
	_ resource.ResourceWithImportState = &assistantResource{}
)

// NewAssistantResource is a helper function to simplify the provider implementation.
func NewAssistantResource() resource.Resource {
	return &assistantResource{}
}

// assistantResource is the resource implementation.
type assistantResource struct {
	client *openai.Client
}

// assistantResourceModel maps the resource schema data.
type assistantResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Description           types.String `tfsdk:"description"`
	Model                 types.String `tfsdk:"model"`
	Instructions          types.String `tfsdk:"instructions"`
	EnableRetrieval       types.Bool   `tfsdk:"enable_retrieval"`
	EnableCodeInterpreter types.Bool   `tfsdk:"enable_code_interpreter"`
	LastUpdated           types.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *assistantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assistant"
}

// Schema defines the schema for the resource.
func (r *assistantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides an OpenAI assistant resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the Assistant.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the assistant.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the assistant.",
				Optional:    true,
			},
			"model": schema.StringAttribute{
				MarkdownDescription: "Model to use for this assistant. Valid options are `gpt-4-turbo-preview`, `gpt-4`, `gpt-3.5-turbo-16k`, `gpt-3.5-turbo-0125`, `gpt-3.5-turbo`, `gpt-4-1106-preview`, `gpt-4-0125-preview`, `gpt-4-0613`, `gpt-3.5-turbo-1106`, `gpt-3.5-turbo-0613` or any other models currently supported by OpenAI assistant.",
				Required:            true,
			},
			"instructions": schema.StringAttribute{
				Description: "Instructions for the assistant. Use this attribute to guide the personality of the assistant and define its goals. Instructions are similar to system messages in the Chat Completions API.",
				Required:    true,
			},
			"enable_retrieval": schema.BoolAttribute{
				Description: "Retrieval enables the assistant with knowledge from files that you or your users upload.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"enable_code_interpreter": schema.BoolAttribute{
				Description: "Code Interpreter enables the assistant to write and run code. This tool can process files with diverse data and formatting, and generate files such as graphs.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the assistant.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *assistantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*openai.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *openai.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Create a new resource.
func (r *assistantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan assistantResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new assistant
	assistantRequest := openai.AssistantRequest{
		Name:         plan.Name.ValueStringPointer(),
		Description:  plan.Description.ValueStringPointer(),
		Model:        plan.Model.ValueString(),
		Instructions: plan.Instructions.ValueStringPointer(),
		Tools:        []openai.AssistantTool{},
	}

	if plan.EnableRetrieval.ValueBool() {
		assistantRequest.Tools = append(assistantRequest.Tools, openai.AssistantTool{Type: openai.AssistantToolTypeRetrieval})
	}

	if plan.EnableCodeInterpreter.ValueBool() {
		assistantRequest.Tools = append(assistantRequest.Tools, openai.AssistantTool{Type: openai.AssistantToolTypeCodeInterpreter})
	}

	assistant, err := r.client.CreateAssistant(ctx, assistantRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating assistant",
			"Could not create assistant, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(assistant.ID)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information.
func (r *assistantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state assistantResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed assistant value from OpenAI
	assistant, err := r.client.RetrieveAssistant(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading OpenAI assistant",
			"Could not read OpenAI assistant ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(assistant.ID)
	state.Name = types.StringValue(*assistant.Name)
	state.Model = types.StringValue(assistant.Model)
	state.Instructions = types.StringValue(*assistant.Instructions)
	state.EnableRetrieval = types.BoolValue(slices.Contains(assistant.Tools, openai.AssistantTool{Type: openai.AssistantToolTypeRetrieval}))
	state.EnableCodeInterpreter = types.BoolValue(slices.Contains(assistant.Tools, openai.AssistantTool{Type: openai.AssistantToolTypeCodeInterpreter}))

	if assistant.Description != nil {
		state.Description = types.StringValue(*assistant.Description)
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *assistantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan assistantResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	assistantRequest := openai.AssistantRequest{
		Name:         plan.Name.ValueStringPointer(),
		Description:  plan.Description.ValueStringPointer(),
		Model:        plan.Model.ValueString(),
		Instructions: plan.Instructions.ValueStringPointer(),
		Tools:        []openai.AssistantTool{},
	}

	if plan.EnableRetrieval.ValueBool() {
		assistantRequest.Tools = append(assistantRequest.Tools, openai.AssistantTool{Type: openai.AssistantToolTypeRetrieval})
	}

	if plan.EnableCodeInterpreter.ValueBool() {
		assistantRequest.Tools = append(assistantRequest.Tools, openai.AssistantTool{Type: openai.AssistantToolTypeCodeInterpreter})
	}

	// Update existing assistant
	_, err := r.client.ModifyAssistant(ctx, plan.ID.ValueString(), assistantRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating OpenAI Assistant",
			"Could not update assistant, unexpected error: "+err.Error(),
		)
		return
	}

	// Fetch updated items from GetAssistant as UpdateAssistant items are not
	// populated.
	_, err = r.client.RetrieveAssistant(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading OpenAI Assistant",
			"Could not read OpenAI assistant ID "+plan.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *assistantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state assistantResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing assistant
	_, err := r.client.DeleteAssistant(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting OpenAI Assistant",
			"Could not delete assistant, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *assistantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
