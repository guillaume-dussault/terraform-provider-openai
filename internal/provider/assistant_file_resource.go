package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	openai "github.com/sashabaranov/go-openai"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &assistantFileResource{}
	_ resource.ResourceWithConfigure   = &assistantFileResource{}
	_ resource.ResourceWithImportState = &assistantFileResource{}
)

// NewAssistantFileResource is a helper function to simplify the provider implementation.
func NewAssistantFileResource() resource.Resource {
	return &assistantFileResource{}
}

// assistantFileResource is the resource implementation.
type assistantFileResource struct {
	client *openai.Client
}

// assistantFileResourceModel maps the resource schema data.
type assistantFileResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Filename    types.String `tfsdk:"filename"`
	AssistantID types.String `tfsdk:"assistant_id"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *assistantFileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assistant_file"
}

// Schema defines the schema for the resource.
func (r *assistantFileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides an OpenAI assistant file resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the file.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"filename": schema.StringAttribute{
				Required:    true,
				Description: "Path to the file within the local filesystem.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"assistant_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the assistant to which this file will be included.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the assistant.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *assistantFileResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *assistantFileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan assistantFileResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	fileContent, err := os.ReadFile(plan.Filename.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading file content",
			"Could not create assistant file, unexpected error: "+err.Error(),
		)
		return
	}

	if len(fileContent) == 0 {
		resp.Diagnostics.AddError(
			"File is empty",
			"Could not create assistant file, the file has no content.",
		)
		return
	}

	name := filepath.Base(plan.Filename.ValueString())

	fileRequest := openai.FileBytesRequest{
		Name:    name,
		Bytes:   fileContent,
		Purpose: "assistants",
	}

	file, err := r.client.CreateFileBytes(ctx, fileRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating file",
			"Could not create assistant file, unexpected error: "+err.Error(),
		)
		return
	}

	_, err = r.client.CreateAssistantFile(ctx, plan.AssistantID.ValueString(), openai.AssistantFileRequest{
		FileID: file.ID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating assistant file",
			"Could not create assistant file, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(file.ID)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information.
func (r *assistantFileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state assistantFileResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed value from OpenAI
	_, err := r.client.GetFile(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading OpenAI file",
			"Could not read OpenAI file ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Get refreshed value from OpenAI
	assistantFile, err := r.client.RetrieveAssistantFile(ctx, state.AssistantID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading OpenAI assistant file",
			"Could not read OpenAI assistant file ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(assistantFile.ID)
	state.AssistantID = types.StringValue(assistantFile.AssistantID)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *assistantFileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan assistantFileResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *assistantFileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state assistantFileResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing assistant file
	err := r.client.DeleteAssistantFile(ctx, state.AssistantID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting OpenAI assistant file",
			"Could not delete assistant file, unexpected error: "+err.Error(),
		)
		return
	}

	// Delete existing file
	err = r.client.DeleteFile(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting OpenAI file",
			"Could not delete assistant file, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *assistantFileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
