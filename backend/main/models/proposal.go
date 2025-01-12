package models

///////////////
// Proposals //
///////////////

import (
	"fmt"
	"os"
	"time"

	s "github.com/brudfyi/flow-voting-tool/main/shared"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

type Proposal struct {
	ID                   int                     `json:"id,omitempty"`
	Name                 string                  `json:"name" validate:"required"`
	Community_id         int                     `json:"communityId"`
	Choices              []string                `json:"choices" validate:"required"`
	Strategy             *string                 `json:"strategy,omitempty"`
	Max_weight           *uint64                 `json:"maxWeight,omitempty"`
	Min_balance          *uint64                 `json:"minBalance,omitempty"`
	Creator_addr         string                  `json:"creatorAddr" validate:"required"`
	Start_time           time.Time               `json:"startTime" validate:"required"`
	Result               *string                 `json:"result,omitempty"`
	End_time             time.Time               `json:"endTime" validate:"required"`
	Created_at           *time.Time              `json:"createdAt,omitempty"`
	Cid                  *string                 `json:"cid,omitempty"`
	Status               *string                 `json:"status,omitempty"`
	Body                 *string                 `json:"body,omitempty" validate:"required"`
	Block_height         uint64                  `json:"block_height"`
	Total_votes          int                     `json:"total_votes"`
	Timestamp            string                  `json:"timestamp" validate:"required"`
	Composite_signatures *[]s.CompositeSignature `json:"compositeSignatures" validate:"required"`
	Computed_status      *string                 `json:"computedStatus,omitempty"`
}

type UpdateProposalRequestPayload struct {
	Status              string                  `json:"status"`
	Timestamp           string                  `json:"timestamp"`
	Composite_signature *[]s.CompositeSignature `json:"compositeSignatures"`
}

var computedStatusSQL = `
	CASE
		WHEN status = 'published' AND start_time > (now() at time zone 'utc') THEN 'pending'
		WHEN status = 'published' AND start_time < (now() at time zone 'utc') AND end_time > (now() at time zone 'utc') THEN 'active'
		WHEN status = 'published' AND end_time < (now() at time zone 'utc') THEN 'closed'
		WHEN status = 'cancelled' THEN 'cancelled'
		WHEN status = 'closed' THEN 'closed'
	END as computed_status
	`

func GetProposalsForCommunity(db *s.Database, start, count int, communityId int, status string, order string) ([]*Proposal, int, error) {
	var proposals []*Proposal
	var err error

	// Get Proposals
	sql := fmt.Sprintf(`SELECT *, %s FROM proposals WHERE community_id = $3`, computedStatusSQL)
	statusFilter := ""

	// Generate SQL based on computed status
	// status: { pending | active | closed | cancelled }
	switch status {
	case "pending":
		statusFilter = ` AND status = 'published' AND start_time > now()`
	case "active":
		statusFilter = ` AND status = 'published' AND start_time < now() AND end_time > now()`
	case "closed":
		statusFilter = ` AND status = 'published' AND end_time < now()`
	case "cancelled":
		statusFilter = ` AND status = 'cancelled'`
	}
	orderBySql := fmt.Sprintf(` ORDER BY created_at %s`, order)
	limitOffsetSql := ` LIMIT $1 OFFSET $2`
	sql = sql + statusFilter + orderBySql + limitOffsetSql

	err = pgxscan.Select(db.Context, db.Conn, &proposals, sql, count, start, communityId)

	// If we get pgx.ErrNoRows, just return an empty array
	// and obfuscate error
	if err != nil && err.Error() != pgx.ErrNoRows.Error() {
		return nil, 0, err
	} else if err != nil && err.Error() == pgx.ErrNoRows.Error() {
		return []*Proposal{}, 0, nil
	}

	// Get total number of proposals
	var totalRecords int
	countSql := `SELECT COUNT(*) FROM proposals WHERE community_id = $1` + statusFilter
	_ = db.Conn.QueryRow(db.Context, countSql, communityId).Scan(&totalRecords)

	return proposals, totalRecords, nil
}

func (p *Proposal) GetProposalById(db *s.Database) error {
	sql := `
	SELECT p.*, %s, count(v.id) as total_votes from proposals as p
	left join votes as v on v.proposal_id = p.id
	WHERE p.id = $1
	GROUP BY p.id`
	sql = fmt.Sprintf(sql, computedStatusSQL)
	return pgxscan.Get(db.Context, db.Conn, p, sql, p.ID)
}

func (p *Proposal) CreateProposal(db *s.Database) error {
	// signaturesJSON := json.Marshal(p.Composite_signatures)
	err := db.Conn.QueryRow(db.Context,
		`
	INSERT INTO proposals(community_id, name, choices, strategy, min_balance, max_weight, creator_addr, start_time, end_time, status, body, block_height, cid, composite_signatures)
	VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	RETURNING id, created_at
	`, p.Community_id, p.Name, p.Choices, p.Strategy, p.Min_balance, p.Max_weight, p.Creator_addr, p.Start_time, p.End_time, p.Status, p.Body, p.Block_height, p.Cid, p.Composite_signatures).Scan(&p.ID, &p.Created_at)

	return err // will be nil unless something went wrong
}

func (p *Proposal) UpdateProposal(db *s.Database) error {
	_, err := db.Conn.Exec(db.Context,
		`
	UPDATE proposals
	SET status = $1
	WHERE id = $2
	`, p.Status, p.ID)

	if err != nil {
		return err
	}

	err = p.GetProposalById(db)
	return err // will be nil unless something went wrong
}

func (p *Proposal) IsLive() bool {
	now := time.Now().UTC()
	return now.After(p.Start_time) && now.Before(p.End_time)
}

// Validations

// Returns an error if the account's balance is insufficient to cast
// a vote on the proposal.
func (p *Proposal) ValidateBalance(balance *Balance) error {
	var weight uint64
	var ERROR error = fmt.Errorf("insufficient balance for strategy: %s", *p.Strategy)

	// TODO: Feature flag
	// Dont validate in DEV or TEST envs!
	if os.Getenv("APP_ENV") == "TEST" || os.Getenv("APP_ENV") == "DEV" {
		return nil
	}

	switch *p.Strategy {
	case "token-weighted-default":
		weight = balance.StakingBalance + balance.PrimaryAccountBalance
	case "staked-token-weighted-default":
		weight = balance.StakingBalance
	case "token-weighted-capped":
		weight = balance.StakingBalance + balance.PrimaryAccountBalance
	default:
		return fmt.Errorf("validateBalance err - Strategy not implemented: %s", *p.Strategy)
	}

	if weight == 0 {
		return ERROR
	}

	if p.Min_balance != nil && *p.Min_balance > 0 && weight < *p.Min_balance {
		return ERROR
	}
	return nil
}
