package diff

import (
	"encoding/json"
)

// JSONPlan represents the JSON output structure for plan
type JSONPlan struct {
	Repo             []JSONChange `json:"repo,omitempty"`
	Topics           []JSONChange `json:"topics,omitempty"`
	Labels           []JSONChange `json:"labels,omitempty"`
	BranchProtection []JSONChange `json:"branch_protection,omitempty"`
	Actions          []JSONChange `json:"actions,omitempty"`
	Pages            []JSONChange `json:"pages,omitempty"`
	Variables        []JSONChange `json:"variables,omitempty"`
	Secrets          []JSONChange `json:"secrets,omitempty"`
	Summary          JSONSummary  `json:"summary"`
}

// JSONChange represents a single change in JSON format
type JSONChange struct {
	Type string      `json:"type"`
	Key  string      `json:"key"`
	Old  interface{} `json:"old,omitempty"`
	New  interface{} `json:"new,omitempty"`
}

// JSONSummary represents the summary counts
type JSONSummary struct {
	Add     int `json:"add"`
	Update  int `json:"update"`
	Delete  int `json:"delete"`
	Missing int `json:"missing"`
}

// ToJSON converts a Plan to JSON output structure
func (p *Plan) ToJSON() *JSONPlan {
	jsonPlan := &JSONPlan{}

	var adds, updates, deletes, missing int

	for _, change := range p.Changes {
		jc := JSONChange{
			Type: change.Type.String(),
			Key:  change.Key,
			Old:  change.Old,
			New:  change.New,
		}

		switch change.Category {
		case "repo":
			jsonPlan.Repo = append(jsonPlan.Repo, jc)
		case "topics":
			jsonPlan.Topics = append(jsonPlan.Topics, jc)
		case "labels":
			jsonPlan.Labels = append(jsonPlan.Labels, jc)
		case "branch_protection":
			jsonPlan.BranchProtection = append(jsonPlan.BranchProtection, jc)
		case "actions":
			jsonPlan.Actions = append(jsonPlan.Actions, jc)
		case "pages":
			jsonPlan.Pages = append(jsonPlan.Pages, jc)
		case "variables":
			jsonPlan.Variables = append(jsonPlan.Variables, jc)
		case "secrets":
			jsonPlan.Secrets = append(jsonPlan.Secrets, jc)
		}

		switch change.Type {
		case ChangeAdd:
			adds++
		case ChangeUpdate:
			updates++
		case ChangeDelete:
			deletes++
		case ChangeMissing:
			missing++
		}
	}

	jsonPlan.Summary = JSONSummary{
		Add:     adds,
		Update:  updates,
		Delete:  deletes,
		Missing: missing,
	}

	return jsonPlan
}

// MarshalIndent returns pretty-printed JSON bytes
func (p *Plan) MarshalIndent() ([]byte, error) {
	return json.MarshalIndent(p.ToJSON(), "", "  ")
}
