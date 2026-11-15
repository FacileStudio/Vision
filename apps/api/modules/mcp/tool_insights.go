package mcp

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/FacileStudio/Vision/apps/api/modules/analytics"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var errComparePair = stderrors.New("compare_from and compare_to go together: pass both, or neither for the period right before")

// listAnomalies is vision_list_anomalies: the stored deviations of a site, as
// GET /analytics/{siteId}/anomalies lists them.
func (t *toolset) listAnomalies(ctx context.Context, req *sdk.CallToolRequest, in anomaliesInput) (*sdk.CallToolResult, anomaliesOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, anomaliesOutput{}, err
	}
	span, err := windowOf(in.From, in.To, time.Now())
	if err != nil {
		return nil, anomaliesOutput{}, err
	}
	site, err := t.site(ctx, userID, in.SiteID)
	if err != nil {
		return nil, anomaliesOutput{}, err
	}
	list, err := t.Analytics.ListAnomalies(ctx, site.ID, analytics.AnomalyQuery{
		From: span.from, To: span.to, Metric: in.Metric, Limit: in.Limit,
	})
	if err != nil {
		return nil, anomaliesOutput{}, mcpkit.Failure(err, "")
	}
	return nil, anomaliesOutput{Site: t.siteOf(site), Range: span.out(), Anomalies: list}, nil
}

// comparePeriods is vision_compare_periods: one period against another, as
// GET /analytics/{siteId}/compare answers.
func (t *toolset) comparePeriods(ctx context.Context, req *sdk.CallToolRequest, in compareInput) (*sdk.CallToolResult, compareOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, compareOutput{}, err
	}
	current, previous, err := periodsOf(in, time.Now())
	if err != nil {
		return nil, compareOutput{}, err
	}
	site, err := t.site(ctx, userID, in.SiteID)
	if err != nil {
		return nil, compareOutput{}, err
	}
	resp, err := t.Analytics.Compare(ctx, site.ID, current, previous, in.Limit)
	if err != nil {
		return nil, compareOutput{}, mcpkit.Failure(err, "")
	}
	return nil, compareOutput{Site: t.siteOf(site), CompareResponse: *resp}, nil
}

// periodsOf reads the period and the one it is compared with, the period of equal
// length right before unless both compare days are given.
func periodsOf(in compareInput, now time.Time) (analytics.Period, analytics.Period, error) {
	span, err := windowOf(in.From, in.To, now)
	if err != nil {
		return analytics.Period{}, analytics.Period{}, err
	}
	current := analytics.Period{From: span.from, To: span.to}
	if in.CompareFrom == "" && in.CompareTo == "" {
		return current, analytics.PreviousPeriod(current), nil
	}
	if in.CompareFrom == "" || in.CompareTo == "" {
		return current, current, errComparePair
	}
	other, err := windowOf(in.CompareFrom, in.CompareTo, now)
	if err != nil {
		return current, current, err
	}
	return current, analytics.Period{From: other.from, To: other.to}, nil
}
