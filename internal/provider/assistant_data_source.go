// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	openai "github.com/sashabaranov/go-openai"
	"golang.org/x/exp/slices"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &assistantDataSource{}
	_ datasource.DataSourceWithConfigure = &assistantDataSource{}
)

// NewAssistantDataSource is a helper function to simplify the provider implementation.
func NewAssistantDataSource() datasource.DataSource {
	return &assistantDataSource{}
}

// assistantDataSource is the data source implementation.
type assistantDataSource struct {
	client *openai.Client
}

// assistantDataSourceModel maps the data source schema data.
type assistantDataSourceModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Description           types.String `tfsdk:"description"`
	Model                 types.String `tfsdk:"model"`
	Instructions          types.String `tfsdk:"instructions"`
	EnableRetrieval       types.Bool   `tfsdk:"enable_retrieval"`
	EnableCodeInterpreter types.Bool   `tfsdk:"enable_code_interpreter"`
}

// Metadata returns the data source type name.
func (d *assistantDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assistant"
}

// Schema defines the schema for the data source.
func (d *assistantDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a OpenAI assistant.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the Assistant.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the assistant.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the assistant.",
				Computed:    true,
			},
			"model": schema.StringAttribute{
				MarkdownDescription: "Model to use for this assistant. Valid options are `gpt-4-turbo-preview`, `gpt-4`, `gpt-3.5-turbo-16k`, `gpt-3.5-turbo-0125`, `gpt-3.5-turbo`, `gpt-4-1106-preview`, `gpt-4-0125-preview`, `gpt-4-0613`, `gpt-3.5-turbo-1106`, `gpt-3.5-turbo-0613` or any other models currently supported by OpenAI assistant.",
				Computed:            true,
			},
			"instructions": schema.StringAttribute{
				Description: "Instructions for the assistant. Use this attribute to guide the personality of the assistant and define its goals. Instructions are similar to system messages in the Chat Completions API.",
				Computed:    true,
			},
			"enable_retrieval": schema.BoolAttribute{
				Description: "Retrieval enables the assistant with knowledge from files that you or your users upload.",
				Computed:    true,
			},
			"enable_code_interpreter": schema.BoolAttribute{
				Description: "Code Interpreter enables the assistant to write and run code. This tool can process files with diverse data and formatting, and generate files such as graphs.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *assistantDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client
}

// Read refreshes the Terraform state with the latest data.
func (d *assistantDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data assistantDataSourceModel

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	assistant, err := d.client.RetrieveAssistant(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read OpenAI assistant",
			err.Error(),
		)
		return
	}

	data.ID = types.StringValue(assistant.ID)
	data.Name = types.StringValue(*assistant.Name)
	data.Model = types.StringValue(assistant.Model)
	data.Instructions = types.StringValue(*assistant.Instructions)
	data.EnableRetrieval = types.BoolValue(slices.Contains(assistant.Tools, openai.AssistantTool{Type: openai.AssistantToolTypeRetrieval}))
	data.EnableCodeInterpreter = types.BoolValue(slices.Contains(assistant.Tools, openai.AssistantTool{Type: openai.AssistantToolTypeCodeInterpreter}))

	if assistant.Description != nil {
		data.Description = types.StringValue(*assistant.Description)
	}

	// Set state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
