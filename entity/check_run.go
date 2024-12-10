package entity

type SimpleCheckRunList struct {
	Total int
	Items []*SimpleCheckRun
}
type SimpleCheckRun struct {
	Name       string
	Status     CheckRunStatus
	Conclusion *CheckRunConclusion
	URL        string
}

type CheckRunStatus string

const (
	CheckRunStatusCompleted CheckRunStatus = "completed"
)

func (c CheckRunStatus) String() string {
	return string(c)
}

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

func (c CheckRunConclusion) String() string {
	return string(c)
}
