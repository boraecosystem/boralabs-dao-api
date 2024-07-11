package error

const (
	EmptyConfigValue     = "Empty value for config key\n"
	InvalidEventName     = "Invalid event name %s\n"
	WrongFromBlockNumber = "Failed convert fromBlock %s\n"
	FailedNewContract    = "Failed new contract :: %v\n"
	FailedParseLogData   = "Failed parse log data :: %v\n"
	FailedSaveLogData    = "Failed save log data :: %v\n"
	FailedUpdateLogData  = "Failed to update log data: %v\n"
	FailedBlockByNumber  = "Failed blockByNumber :: %v\n"
	FailedExistsProposal = "This is an proposal that already exists."
)
