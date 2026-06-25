package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
	"github.com/google/uuid"
	"github.com/oapi-codegen/nullable"
	"github.com/spf13/cobra"
)

// RAG deploy flags.
var (
	ragDeployEmbeddingLLMProvider  string
	ragDeployEmbeddingLLMModel     string
	ragDeployEmbeddingLLMAPIKey    string
	ragDeployCompletionLLMProvider string
	ragDeployCompletionLLMModel    string
	ragDeployCompletionLLMAPIKey   string
	ragDeployTokenBudget           int
	ragDeployTopN                  int
	ragDeployPipelineConfig        string
	ragDeployTargetNodes           []string
)

// RAG update flags.
var (
	ragUpdateEmbeddingLLMProvider  string
	ragUpdateEmbeddingLLMModel     string
	ragUpdateEmbeddingLLMAPIKey    string
	ragUpdateCompletionLLMProvider string
	ragUpdateCompletionLLMModel    string
	ragUpdateCompletionLLMAPIKey   string
	ragUpdateTokenBudget           int
	ragUpdateTopN                  int
	ragUpdatePipelineConfig        string
	ragUpdateTargetNodes           []string
)

func init() {
	databasesCmd.AddCommand(dbRAGCmd)
	dbRAGCmd.AddCommand(dbRAGDeployCmd)
	dbRAGCmd.AddCommand(dbRAGUpdateCmd)

	// deploy flags
	dbRAGDeployCmd.Flags().StringVar(&ragDeployEmbeddingLLMProvider,
		"embedding-llm-provider", "",
		"Embedding LLM provider (e.g. openai, voyage)")
	dbRAGDeployCmd.Flags().StringVar(&ragDeployEmbeddingLLMModel,
		"embedding-llm-model", "",
		"Embedding LLM model identifier")
	dbRAGDeployCmd.Flags().StringVar(&ragDeployEmbeddingLLMAPIKey,
		"embedding-llm-api-key", "",
		"API key for the embedding LLM provider")
	dbRAGDeployCmd.Flags().StringVar(&ragDeployCompletionLLMProvider,
		"completion-llm-provider", "",
		"Completion LLM provider (e.g. openai)")
	dbRAGDeployCmd.Flags().StringVar(&ragDeployCompletionLLMModel,
		"completion-llm-model", "",
		"Completion LLM model identifier")
	dbRAGDeployCmd.Flags().StringVar(&ragDeployCompletionLLMAPIKey,
		"completion-llm-api-key", "",
		"API key for the completion LLM provider")
	dbRAGDeployCmd.Flags().IntVar(&ragDeployTokenBudget,
		"token-budget", 0,
		"Default max completion tokens across all pipelines")
	dbRAGDeployCmd.Flags().IntVar(&ragDeployTopN,
		"top-n", 0,
		"Default number of results to retrieve per pipeline")
	dbRAGDeployCmd.Flags().StringVar(&ragDeployPipelineConfig,
		"pipeline-config", "",
		"Path to a JSON file containing pipeline definitions")
	dbRAGDeployCmd.Flags().StringSliceVar(&ragDeployTargetNodes,
		"target-nodes", nil,
		"Node names to deploy on (e.g. n1,n2). Auto-selects if cluster has one node")
	addServiceWaitFlags(dbRAGDeployCmd)

	_ = dbRAGDeployCmd.MarkFlagRequired("embedding-llm-provider")
	_ = dbRAGDeployCmd.MarkFlagRequired("embedding-llm-model")
	_ = dbRAGDeployCmd.MarkFlagRequired("embedding-llm-api-key")
	_ = dbRAGDeployCmd.MarkFlagRequired("completion-llm-provider")
	_ = dbRAGDeployCmd.MarkFlagRequired("completion-llm-model")
	_ = dbRAGDeployCmd.MarkFlagRequired("completion-llm-api-key")
	_ = dbRAGDeployCmd.MarkFlagRequired("pipeline-config")

	// update flags (same set, none required)
	dbRAGUpdateCmd.Flags().StringVar(&ragUpdateEmbeddingLLMProvider,
		"embedding-llm-provider", "",
		"Embedding LLM provider (e.g. openai, voyage)")
	dbRAGUpdateCmd.Flags().StringVar(&ragUpdateEmbeddingLLMModel,
		"embedding-llm-model", "",
		"Embedding LLM model identifier")
	dbRAGUpdateCmd.Flags().StringVar(&ragUpdateEmbeddingLLMAPIKey,
		"embedding-llm-api-key", "",
		"API key for the embedding LLM provider")
	dbRAGUpdateCmd.Flags().StringVar(&ragUpdateCompletionLLMProvider,
		"completion-llm-provider", "",
		"Completion LLM provider (e.g. openai)")
	dbRAGUpdateCmd.Flags().StringVar(&ragUpdateCompletionLLMModel,
		"completion-llm-model", "",
		"Completion LLM model identifier")
	dbRAGUpdateCmd.Flags().StringVar(&ragUpdateCompletionLLMAPIKey,
		"completion-llm-api-key", "",
		"API key for the completion LLM provider")
	dbRAGUpdateCmd.Flags().IntVar(&ragUpdateTokenBudget,
		"token-budget", 0,
		"Default max completion tokens across all pipelines")
	dbRAGUpdateCmd.Flags().IntVar(&ragUpdateTopN,
		"top-n", 0,
		"Default number of results to retrieve per pipeline")
	dbRAGUpdateCmd.Flags().StringVar(&ragUpdatePipelineConfig,
		"pipeline-config", "",
		"Path to a JSON file containing pipeline definitions")
	dbRAGUpdateCmd.Flags().StringSliceVar(&ragUpdateTargetNodes,
		"target-nodes", nil,
		"Node names to deploy on (e.g. n1,n2). Auto-selects if cluster has one node")
	addServiceWaitFlags(dbRAGUpdateCmd)
}

var dbRAGCmd = &cobra.Command{
	Use:   "rag",
	Short: "Manage the RAG server deployed on a database",
}

// --- deploy ---

var dbRAGDeployCmd = &cobra.Command{
	Use:   "deploy <db-id>",
	Short: "Deploy a RAG server alongside a database",
	Args:  cobra.ExactArgs(1),
	RunE:  runDBRAGDeploy,
}

func runDBRAGDeploy(cmd *cobra.Command, args []string) error {
	return applyRAGService(cmd, args[0],
		ragDeployEmbeddingLLMProvider,
		ragDeployEmbeddingLLMModel,
		ragDeployEmbeddingLLMAPIKey,
		ragDeployCompletionLLMProvider,
		ragDeployCompletionLLMModel,
		ragDeployCompletionLLMAPIKey,
		ragDeployTokenBudget,
		ragDeployTopN,
		ragDeployPipelineConfig,
		ragDeployTargetNodes,
	)
}

// --- update ---

var dbRAGUpdateCmd = &cobra.Command{
	Use:   "update <db-id>",
	Short: "Update the RAG server configuration on a database",
	Args:  cobra.ExactArgs(1),
	RunE:  runDBRAGUpdate,
}

func runDBRAGUpdate(cmd *cobra.Command, args []string) error {
	return applyRAGService(cmd, args[0],
		ragUpdateEmbeddingLLMProvider,
		ragUpdateEmbeddingLLMModel,
		ragUpdateEmbeddingLLMAPIKey,
		ragUpdateCompletionLLMProvider,
		ragUpdateCompletionLLMModel,
		ragUpdateCompletionLLMAPIKey,
		ragUpdateTokenBudget,
		ragUpdateTopN,
		ragUpdatePipelineConfig,
		ragUpdateTargetNodes,
	)
}

// --- shared implementation ---

func applyRAGService(
	cmd *cobra.Command,
	dbID string,
	embeddingProvider string,
	embeddingModel string,
	embeddingAPIKey string,
	completionProvider string,
	completionModel string,
	completionAPIKey string,
	tokenBudget int,
	topN int,
	pipelineConfigPath string,
	targetNodes []string,
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

	// Build embedding LLM config if provider/model supplied.
	embeddingLLM := api.RAGLLMConfig{}
	if embeddingProvider != "" {
		embeddingLLM.Provider = api.RAGLLMConfigProvider(embeddingProvider)
	}
	if embeddingModel != "" {
		embeddingLLM.Model = embeddingModel
	}
	if embeddingAPIKey != "" {
		embeddingLLM.ApiKey = &embeddingAPIKey
	}

	// Build completion LLM config if provider/model supplied.
	completionLLM := api.RAGLLMConfig{}
	if completionProvider != "" {
		completionLLM.Provider = api.RAGLLMConfigProvider(completionProvider)
	}
	if completionModel != "" {
		completionLLM.Model = completionModel
	}
	if completionAPIKey != "" {
		completionLLM.ApiKey = &completionAPIKey
	}

	// Read and parse pipeline config file if provided.
	var pipelines []api.RAGPipelineConfig
	if pipelineConfigPath != "" {
		data, readErr := os.ReadFile(pipelineConfigPath)
		if readErr != nil {
			return fmt.Errorf("read pipeline config %q: %w", pipelineConfigPath, readErr)
		}
		pipelines, err = parsePipelineConfig(data)
		if err != nil {
			return fmt.Errorf("parse pipeline config %q: %w",
				pipelineConfigPath, err)
		}
	}

	ragCfg := &api.RAGServiceConfig{
		EmbeddingLlm:  embeddingLLM,
		CompletionLlm: completionLLM,
		Pipelines:     pipelines,
	}

	if tokenBudget > 0 {
		ragCfg.TokenBudget = &tokenBudget
	}
	if topN > 0 {
		ragCfg.TopN = &topN
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
		ServiceType: api.ServiceConfigServiceTypeRag,
		RagConfig:   ragCfg,
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
		return fmt.Errorf("apply RAG service: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "RAG service applied to database %s.\n", dbID)
	return trackServiceMutation(cmd, client, dbID, priorTaskID)
}

// parsePipelineConfig parses a RAG pipeline definition file. It accepts either
// a bare JSON array of pipelines or an object with a "pipelines" key — the
// latter mirrors the shape the API returns under rag_config, so a config read
// back from the API can be pasted into a file and used directly.
func parsePipelineConfig(data []byte) ([]api.RAGPipelineConfig, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && trimmed[0] == '{' {
		// Object form: require an explicit "pipelines" key so a typo such as
		// {"pipeline": [...]} is rejected rather than silently treated as an
		// empty pipeline list. An explicit {"pipelines": []} is allowed.
		var probe map[string]json.RawMessage
		if err := json.Unmarshal(trimmed, &probe); err != nil {
			return nil, err
		}
		raw, ok := probe["pipelines"]
		if !ok {
			return nil, fmt.Errorf(`object form must contain a "pipelines" key`)
		}
		var pipelines []api.RAGPipelineConfig
		if err := json.Unmarshal(raw, &pipelines); err != nil {
			return nil, err
		}
		return pipelines, nil
	}

	var pipelines []api.RAGPipelineConfig
	if err := json.Unmarshal(trimmed, &pipelines); err != nil {
		return nil, err
	}
	return pipelines, nil
}
