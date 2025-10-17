package workflow

type WorkflowResult string

const (
	//Status
	New        WorkflowResult = "NEW"
	Processing WorkflowResult = "PROCESSING"
	Completed  WorkflowResult = "COMPLETED"

	//Step
	Skip    WorkflowResult = "SKIP"
	Failed  WorkflowResult = "FAILED"
	Succeed WorkflowResult = "SUCCEED"
	Rerun   WorkflowResult = "RERUN"
)

const WF_DEFAULT_TIMEOUT = 10000
