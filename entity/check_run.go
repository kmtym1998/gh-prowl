package entity

type SimpleCheckRunList struct {
	Total int
	Items []*SimpleCheckRun
}
type SimpleCheckRun struct {
	Name       string              `json:"name"`
	Status     CheckRunStatus      `json:"status"`
	Conclusion *CheckRunConclusion `json:"conclusion"`
}

type CheckRunStatus string

const (
	CheckRunStatusCompleted CheckRunStatus = "completed"
)

type CheckRunConclusion string

const (
	CheckRunConclusionSuccess        CheckRunConclusion = "success"
	CheckRunConclusionFailure        CheckRunConclusion = "failure"
	CheckRunConclusionNeutral        CheckRunConclusion = "neutral"
	CheckRunConclusionCancelled      CheckRunConclusion = "cancelled"
	CheckRunConclusionSkipped        CheckRunConclusion = "skipped"
	CheckRunConclusionTimedOut       CheckRunConclusion = "timed_out"
	CheckRunConclusionActionRequired CheckRunConclusion = "action_required"
)

func (c CheckRunConclusion) IsSuccess() bool {
	return c == CheckRunConclusionSuccess
}
