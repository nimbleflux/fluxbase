package config

// Merge functions for tenant configuration overrides

// mergeAuthConfig merges auth overrides with base auth config
func mergeAuthConfig(base AuthConfig, override AuthConfig) AuthConfig {
	merged := *deepCopyAuthConfig(&base)

	if override.JWTSecret != "" {
		merged.JWTSecret = override.JWTSecret
	}
	if override.JWTExpiry != 0 {
		merged.JWTExpiry = override.JWTExpiry
	}
	if override.RefreshExpiry != 0 {
		merged.RefreshExpiry = override.RefreshExpiry
	}
	if override.ServiceRoleTTL != 0 {
		merged.ServiceRoleTTL = override.ServiceRoleTTL
	}
	if override.AnonTTL != 0 {
		merged.AnonTTL = override.AnonTTL
	}
	if override.MagicLinkExpiry != 0 {
		merged.MagicLinkExpiry = override.MagicLinkExpiry
	}
	if override.PasswordResetExpiry != 0 {
		merged.PasswordResetExpiry = override.PasswordResetExpiry
	}
	if override.PasswordMinLen != 0 {
		merged.PasswordMinLen = override.PasswordMinLen
	}
	if override.BcryptCost != 0 {
		merged.BcryptCost = override.BcryptCost
	}
	if override.TOTPIssuer != "" {
		merged.TOTPIssuer = override.TOTPIssuer
	}
	if override.OAuthProviders != nil {
		merged.OAuthProviders = override.OAuthProviders
	}
	if override.SAMLProviders != nil {
		merged.SAMLProviders = override.SAMLProviders
	}

	return merged
}

// deepCopyAuthConfig creates a deep copy of AuthConfig
func deepCopyAuthConfig(src *AuthConfig) *AuthConfig {
	if src == nil {
		return &AuthConfig{}
	}
	cpy := *src
	if src.OAuthProviders != nil {
		cpy.OAuthProviders = make([]OAuthProviderConfig, len(src.OAuthProviders))
		copy(cpy.OAuthProviders, src.OAuthProviders)
	}
	if src.SAMLProviders != nil {
		cpy.SAMLProviders = make([]SAMLProviderConfig, len(src.SAMLProviders))
		copy(cpy.SAMLProviders, src.SAMLProviders)
	}
	return &cpy
}

// mergeStorageConfig merges storage overrides with base storage config
func mergeStorageConfig(base StorageConfig, override StorageConfig) StorageConfig {
	merged := *deepCopyStorageConfig(&base)

	if override.Provider != "" {
		merged.Provider = override.Provider
	}
	if override.LocalPath != "" {
		merged.LocalPath = override.LocalPath
	}
	if override.S3Bucket != "" {
		merged.S3Bucket = override.S3Bucket
	}
	if override.S3Region != "" {
		merged.S3Region = override.S3Region
	}
	if override.S3Endpoint != "" {
		merged.S3Endpoint = override.S3Endpoint
	}
	if override.S3AccessKey != "" {
		merged.S3AccessKey = override.S3AccessKey
	}
	if override.S3SecretKey != "" {
		merged.S3SecretKey = override.S3SecretKey
	}
	if override.MaxUploadSize != 0 {
		merged.MaxUploadSize = override.MaxUploadSize
	}
	if override.DefaultBuckets != nil {
		merged.DefaultBuckets = override.DefaultBuckets
	}

	return merged
}

// deepCopyStorageConfig creates a deep copy of StorageConfig
func deepCopyStorageConfig(src *StorageConfig) *StorageConfig {
	if src == nil {
		return &StorageConfig{}
	}
	cpy := *src
	if src.DefaultBuckets != nil {
		cpy.DefaultBuckets = make([]string, len(src.DefaultBuckets))
		copy(cpy.DefaultBuckets, src.DefaultBuckets)
	}
	return &cpy
}

// mergeEmailConfig merges email overrides with base email config
func mergeEmailConfig(base EmailConfig, override EmailConfig) EmailConfig {
	merged := *deepCopyEmailConfig(&base)

	if override.Provider != "" {
		merged.Provider = override.Provider
	}
	if override.FromAddress != "" {
		merged.FromAddress = override.FromAddress
	}
	if override.FromName != "" {
		merged.FromName = override.FromName
	}
	if override.ReplyToAddress != "" {
		merged.ReplyToAddress = override.ReplyToAddress
	}
	if override.SMTPHost != "" {
		merged.SMTPHost = override.SMTPHost
	}
	if override.SMTPPort != 0 {
		merged.SMTPPort = override.SMTPPort
	}
	if override.SMTPUsername != "" {
		merged.SMTPUsername = override.SMTPUsername
	}
	if override.SMTPPassword != "" {
		merged.SMTPPassword = override.SMTPPassword
	}
	if override.SendGridAPIKey != "" {
		merged.SendGridAPIKey = override.SendGridAPIKey
	}
	if override.MailgunAPIKey != "" {
		merged.MailgunAPIKey = override.MailgunAPIKey
	}
	if override.MailgunDomain != "" {
		merged.MailgunDomain = override.MailgunDomain
	}
	if override.SESAccessKey != "" {
		merged.SESAccessKey = override.SESAccessKey
	}
	if override.SESSecretKey != "" {
		merged.SESSecretKey = override.SESSecretKey
	}
	if override.SESRegion != "" {
		merged.SESRegion = override.SESRegion
	}

	return merged
}

// deepCopyEmailConfig creates a deep copy of EmailConfig
func deepCopyEmailConfig(src *EmailConfig) *EmailConfig {
	if src == nil {
		return &EmailConfig{}
	}
	cpy := *src
	return &cpy
}

// mergeFunctionsConfig merges functions overrides with base functions config
func mergeFunctionsConfig(base FunctionsConfig, override FunctionsConfig) FunctionsConfig {
	merged := *deepCopyFunctionsConfig(&base)

	if override.FunctionsDir != "" {
		merged.FunctionsDir = override.FunctionsDir
	}
	if override.DefaultTimeout != 0 {
		merged.DefaultTimeout = override.DefaultTimeout
	}
	if override.MaxTimeout != 0 {
		merged.MaxTimeout = override.MaxTimeout
	}
	if override.DefaultMemoryLimit != 0 {
		merged.DefaultMemoryLimit = override.DefaultMemoryLimit
	}
	if override.MaxMemoryLimit != 0 {
		merged.MaxMemoryLimit = override.MaxMemoryLimit
	}

	return merged
}

// deepCopyFunctionsConfig creates a deep copy of FunctionsConfig
func deepCopyFunctionsConfig(src *FunctionsConfig) *FunctionsConfig {
	if src == nil {
		return &FunctionsConfig{}
	}
	cpy := *src
	return &cpy
}

// mergeJobsConfig merges jobs overrides with base jobs config
func mergeJobsConfig(base JobsConfig, override JobsConfig) JobsConfig {
	merged := *deepCopyJobsConfig(&base)

	if override.JobsDir != "" {
		merged.JobsDir = override.JobsDir
	}
	if override.EmbeddedWorkerCount != 0 {
		merged.EmbeddedWorkerCount = override.EmbeddedWorkerCount
	}
	if override.MaxConcurrentPerWorker != 0 {
		merged.MaxConcurrentPerWorker = override.MaxConcurrentPerWorker
	}
	if override.DefaultMaxDuration != 0 {
		merged.DefaultMaxDuration = override.DefaultMaxDuration
	}
	if override.PollInterval != 0 {
		merged.PollInterval = override.PollInterval
	}

	return merged
}

// deepCopyJobsConfig creates a deep copy of JobsConfig
func deepCopyJobsConfig(src *JobsConfig) *JobsConfig {
	if src == nil {
		return &JobsConfig{}
	}
	cpy := *src
	return &cpy
}

// mergeAIConfig merges AI overrides with base AI config
func mergeAIConfig(base AIConfig, override AIConfig) AIConfig {
	merged := *deepCopyAIConfig(&base)

	if override.ChatbotsDir != "" {
		merged.ChatbotsDir = override.ChatbotsDir
	}
	if override.DefaultMaxTokens != 0 {
		merged.DefaultMaxTokens = override.DefaultMaxTokens
	}
	if override.DefaultModel != "" {
		merged.DefaultModel = override.DefaultModel
	}
	if override.ProviderType != "" {
		merged.ProviderType = override.ProviderType
	}
	if override.ProviderModel != "" {
		merged.ProviderModel = override.ProviderModel
	}
	if override.EmbeddingProvider != "" {
		merged.EmbeddingProvider = override.EmbeddingProvider
	}

	return merged
}

// deepCopyAIConfig creates a deep copy of AIConfig
func deepCopyAIConfig(src *AIConfig) *AIConfig {
	if src == nil {
		return &AIConfig{}
	}
	cpy := *src
	return &cpy
}

// mergeRealtimeConfig merges realtime overrides with base realtime config
func mergeRealtimeConfig(base RealtimeConfig, override RealtimeConfig) RealtimeConfig {
	merged := *deepCopyRealtimeConfig(&base)

	if override.MaxConnections != 0 {
		merged.MaxConnections = override.MaxConnections
	}
	if override.MaxConnectionsPerUser != 0 {
		merged.MaxConnectionsPerUser = override.MaxConnectionsPerUser
	}
	if override.PingInterval != 0 {
		merged.PingInterval = override.PingInterval
	}

	return merged
}

// deepCopyRealtimeConfig creates a deep copy of RealtimeConfig
func deepCopyRealtimeConfig(src *RealtimeConfig) *RealtimeConfig {
	if src == nil {
		return &RealtimeConfig{}
	}
	cpy := *src
	return &cpy
}

// mergeAPIConfig merges API overrides with base API config
func mergeAPIConfig(base APIConfig, override APIConfig) APIConfig {
	merged := *deepCopyAPIConfig(&base)

	if override.MaxPageSize != 0 {
		merged.MaxPageSize = override.MaxPageSize
	}
	if override.MaxTotalResults != 0 {
		merged.MaxTotalResults = override.MaxTotalResults
	}
	if override.DefaultPageSize != 0 {
		merged.DefaultPageSize = override.DefaultPageSize
	}
	if override.MaxBatchSize != 0 {
		merged.MaxBatchSize = override.MaxBatchSize
	}

	return merged
}

// deepCopyAPIConfig creates a deep copy of APIConfig
func deepCopyAPIConfig(src *APIConfig) *APIConfig {
	if src == nil {
		return &APIConfig{}
	}
	cpy := *src
	return &cpy
}

// mergeGraphQLConfig merges GraphQL overrides with base GraphQL config
func mergeGraphQLConfig(base GraphQLConfig, override GraphQLConfig) GraphQLConfig {
	merged := *deepCopyGraphQLConfig(&base)

	if override.MaxDepth != 0 {
		merged.MaxDepth = override.MaxDepth
	}
	if override.MaxComplexity != 0 {
		merged.MaxComplexity = override.MaxComplexity
	}

	return merged
}

// deepCopyGraphQLConfig creates a deep copy of GraphQLConfig
func deepCopyGraphQLConfig(src *GraphQLConfig) *GraphQLConfig {
	if src == nil {
		return &GraphQLConfig{}
	}
	cpy := *src
	return &cpy
}

// mergeRPCConfig merges RPC overrides with base RPC config
func mergeRPCConfig(base RPCConfig, override RPCConfig) RPCConfig {
	merged := *deepCopyRPCConfig(&base)

	if override.ProceduresDir != "" {
		merged.ProceduresDir = override.ProceduresDir
	}
	if override.DefaultMaxExecutionTime != 0 {
		merged.DefaultMaxExecutionTime = override.DefaultMaxExecutionTime
	}
	if override.DefaultMaxRows != 0 {
		merged.DefaultMaxRows = override.DefaultMaxRows
	}

	return merged
}

// deepCopyRPCConfig creates a deep copy of RPCConfig
func deepCopyRPCConfig(src *RPCConfig) *RPCConfig {
	if src == nil {
		return &RPCConfig{}
	}
	cpy := *src
	return &cpy
}
