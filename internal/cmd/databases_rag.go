package cmd

import (
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
		"Ordered list of database node names the RAG service connects to")

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
		"Ordered list of database node names the RAG service connects to")
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
	db, err := fetchDatabase(dbID)
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

	// Read and unmarshal pipeline config file if provided.
	var pipelines []api.RAGPipelineConfig
	if pipelineConfigPath != "" {
		data, readErr := os.ReadFile(pipelineConfigPath)
		if readErr != nil {
			return fmt.Errorf("read pipeline config %q: %w", pipelineConfigPath, readErr)
		}
		if unmarshalErr := json.Unmarshal(data, &pipelines); unmarshalErr != nil {
			return fmt.Errorf("parse pipeline config %q: %w",
				pipelineConfigPath, unmarshalErr)
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

	newSvc := api.ServiceConfig{
		ServiceType: api.ServiceConfigServiceTypeRag,
		RagConfig:   ragCfg,
	}

	if len(targetNodes) > 0 {
		newSvc.TargetNodes = &targetNodes
	}

	services := buildServiceList(db, newSvc)
	svcs := nullable.NewNullableWithValue(services)

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	body := api.UpdateDatabaseJSONRequestBody{
		Services: svcs,
	}

	resp, err := client.UpdateDatabaseWithResponse(context.Background(), id, body)
	if err != nil {
		return fmt.Errorf("apply RAG service: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "RAG service applied to database %s.\n", dbID)
	return nil
}
