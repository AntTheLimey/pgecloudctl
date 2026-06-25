package cmd

import (
	"context"
	"fmt"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
	"github.com/google/uuid"
	"github.com/oapi-codegen/nullable"
	"github.com/spf13/cobra"
)

// MCP deploy flags.
var (
	mcpDeployAllowWrites       bool
	mcpDeployEmbeddingProvider string
	mcpDeployEmbeddingModel    string
	mcpDeployEmbeddingAPIKey   string
	mcpDeployOllamaURL         string
	mcpDeployTargetNodes       []string
	mcpDeployInitTokens        string
	mcpDeployInitUsers         string
)

// MCP update flags (mirrors deploy).
var (
	mcpUpdateAllowWrites       bool
	mcpUpdateEmbeddingProvider string
	mcpUpdateEmbeddingModel    string
	mcpUpdateEmbeddingAPIKey   string
	mcpUpdateOllamaURL         string
	mcpUpdateTargetNodes       []string
	mcpUpdateInitTokens        string
	mcpUpdateInitUsers         string
)

func init() {
	databasesCmd.AddCommand(dbMCPCmd)
	dbMCPCmd.AddCommand(dbMCPDeployCmd)
	dbMCPCmd.AddCommand(dbMCPUpdateCmd)

	// deploy flags
	dbMCPDeployCmd.Flags().BoolVar(&mcpDeployAllowWrites, "allow-writes", false,
		"Grant the MCP service read-write access (WARNING: allows LLM to modify data)")
	dbMCPDeployCmd.Flags().StringVar(&mcpDeployEmbeddingProvider, "embedding-provider", "",
		"Embedding provider: ollama, openai, or voyage")
	dbMCPDeployCmd.Flags().StringVar(&mcpDeployEmbeddingModel, "embedding-model", "",
		"Embedding model identifier (required when --embedding-provider is set)")
	dbMCPDeployCmd.Flags().StringVar(&mcpDeployEmbeddingAPIKey, "embedding-api-key", "",
		"API key for the embedding provider (required for openai and voyage)")
	dbMCPDeployCmd.Flags().StringVar(&mcpDeployOllamaURL, "ollama-url", "",
		"Endpoint URL for an Ollama server (required when --embedding-provider is ollama)")
	dbMCPDeployCmd.Flags().StringSliceVar(&mcpDeployTargetNodes, "target-nodes", nil,
		"Node names to deploy on (e.g. n1,n2). Auto-selects if cluster has one node")
	dbMCPDeployCmd.Flags().StringVar(&mcpDeployInitTokens, "init-tokens", "",
		"Bearer token forwarded to the MCP server as INIT_TOKENS")
	dbMCPDeployCmd.Flags().StringVar(&mcpDeployInitUsers, "init-users", "",
		"Comma-separated username:password pairs forwarded as INIT_USERS")
	addServiceWaitFlags(dbMCPDeployCmd)

	// update flags (same set)
	dbMCPUpdateCmd.Flags().BoolVar(&mcpUpdateAllowWrites, "allow-writes", false,
		"Grant the MCP service read-write access (WARNING: allows LLM to modify data)")
	dbMCPUpdateCmd.Flags().StringVar(&mcpUpdateEmbeddingProvider, "embedding-provider", "",
		"Embedding provider: ollama, openai, or voyage")
	dbMCPUpdateCmd.Flags().StringVar(&mcpUpdateEmbeddingModel, "embedding-model", "",
		"Embedding model identifier (required when --embedding-provider is set)")
	dbMCPUpdateCmd.Flags().StringVar(&mcpUpdateEmbeddingAPIKey, "embedding-api-key", "",
		"API key for the embedding provider (required for openai and voyage)")
	dbMCPUpdateCmd.Flags().StringVar(&mcpUpdateOllamaURL, "ollama-url", "",
		"Endpoint URL for an Ollama server (required when --embedding-provider is ollama)")
	dbMCPUpdateCmd.Flags().StringSliceVar(&mcpUpdateTargetNodes, "target-nodes", nil,
		"Node names to deploy on (e.g. n1,n2). Auto-selects if cluster has one node")
	dbMCPUpdateCmd.Flags().StringVar(&mcpUpdateInitTokens, "init-tokens", "",
		"Bearer token forwarded to the MCP server as INIT_TOKENS")
	dbMCPUpdateCmd.Flags().StringVar(&mcpUpdateInitUsers, "init-users", "",
		"Comma-separated username:password pairs forwarded as INIT_USERS")
	addServiceWaitFlags(dbMCPUpdateCmd)
}

var dbMCPCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage the MCP server deployed on a database",
}

// --- deploy ---

var dbMCPDeployCmd = &cobra.Command{
	Use:   "deploy <db-id>",
	Short: "Deploy an MCP server alongside a database",
	Args:  cobra.ExactArgs(1),
	RunE:  runDBMCPDeploy,
}

func runDBMCPDeploy(cmd *cobra.Command, args []string) error {
	return applyMCPService(cmd, args[0],
		mcpDeployAllowWrites,
		mcpDeployEmbeddingProvider,
		mcpDeployEmbeddingModel,
		mcpDeployEmbeddingAPIKey,
		mcpDeployOllamaURL,
		mcpDeployTargetNodes,
		mcpDeployInitTokens,
		mcpDeployInitUsers,
	)
}

// --- update ---

var dbMCPUpdateCmd = &cobra.Command{
	Use:   "update <db-id>",
	Short: "Update the MCP server configuration on a database",
	Args:  cobra.ExactArgs(1),
	RunE:  runDBMCPUpdate,
}

func runDBMCPUpdate(cmd *cobra.Command, args []string) error {
	return applyMCPService(cmd, args[0],
		mcpUpdateAllowWrites,
		mcpUpdateEmbeddingProvider,
		mcpUpdateEmbeddingModel,
		mcpUpdateEmbeddingAPIKey,
		mcpUpdateOllamaURL,
		mcpUpdateTargetNodes,
		mcpUpdateInitTokens,
		mcpUpdateInitUsers,
	)
}

// --- shared implementation ---

func applyMCPService(
	cmd *cobra.Command,
	dbID string,
	allowWrites bool,
	embeddingProvider string,
	embeddingModel string,
	embeddingAPIKey string,
	ollamaURL string,
	targetNodes []string,
	initTokens string,
	initUsers string,
) error {
	client, db, err := fetchDatabase(dbID)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(dbID)
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid database ID %q: %v", dbID, err),
			code: ExitGeneral,
		}
	}

	mcpCfg := &api.MCPServiceConfig{}
	mcpCfg.AllowWrites = &allowWrites

	if embeddingProvider != "" {
		p := api.MCPServiceConfigEmbeddingProvider(embeddingProvider)
		mcpCfg.EmbeddingProvider = &p
	}
	if embeddingModel != "" {
		mcpCfg.EmbeddingModel = &embeddingModel
	}
	if embeddingAPIKey != "" {
		mcpCfg.EmbeddingApiKey = &embeddingAPIKey
	}
	if ollamaURL != "" {
		mcpCfg.OllamaUrl = &ollamaURL
	}
	if initTokens != "" {
		mcpCfg.InitTokens = &initTokens
	}
	if initUsers != "" {
		mcpCfg.InitUsers = &initUsers
	}

	clusterID, err := uuid.Parse(db.ClusterId)
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid cluster ID %q on database: %v", db.ClusterId, err),
			code: ExitGeneral,
		}
	}

	hostIDs, err := resolveHostIDs(client, clusterID, targetNodes)
	if err != nil {
		return err
	}

	newSvc := api.ServiceConfig{
		ServiceType: api.ServiceConfigServiceTypeMcp,
		McpConfig:   mcpCfg,
		HostIds:     &hostIDs,
	}

	services := buildServiceList(db, newSvc)
	svcs := nullable.NewNullableWithValue(services)

	body := api.UpdateDatabaseJSONRequestBody{
		Services: svcs,
	}

	var priorTaskID string
	if svcWait {
		priorTaskID, err = newestSubjectTaskID(context.Background(), client, dbID)
		if err != nil {
			return err
		}
	}

	resp, err := client.UpdateDatabaseWithResponse(context.Background(), id, body)
	if err != nil {
		return fmt.Errorf("apply MCP service: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "MCP service applied to database %s.\n", dbID)
	return trackServiceMutation(cmd, client, dbID, priorTaskID)
}
