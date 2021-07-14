package pullrequest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
)

func TestGitlab(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	server := spawnTestServer()
	defer server.Close()

	cfg := &config.Config{}
	cfg.Gitlab.Host = server.URL
	cfg.Gitlab.AccessToken = "https://gitlab.example.com"
	cfg.Gitlab.AccessToken = "0815"

	commands := bot.Commands{}
	cmd := newGitlabCommand(base, cfg).(command)
	gitlabFetcher := cmd.fetcher.(*gitlabFetcher)

	commands.AddCommand(cmd)

	t.Run("help", func(t *testing.T) {
		help := commands.GetHelp()
		assert.Equal(t, 1, len(help))
	})

	t.Run("get status", func(t *testing.T) {
		mr := &gitlab.MergeRequest{}
		actual := gitlabFetcher.getStatus(mr)
		assert.Equal(t, prStatusOpen, actual)

		mr = &gitlab.MergeRequest{}
		mr.State = "merged"
		actual = gitlabFetcher.getStatus(mr)
		assert.Equal(t, prStatusMerged, actual)

		mr = &gitlab.MergeRequest{}
		mr.State = "closed"
		actual = gitlabFetcher.getStatus(mr)
		assert.Equal(t, prStatusClosed, actual)
	})

	t.Run("test convertToPullRequest", func(t *testing.T) {
		mr := &gitlab.MergeRequest{}
		mr.State = "open"
		mr.Pipeline = &gitlab.PipelineInfo{}
		mr.Pipeline.Status = "running"
		mr.Title = "my title"

		actual := gitlabFetcher.convertToPullRequest(mr, 100)

		expected := pullRequest{}
		expected.Status = prStatusOpen
		expected.BuildStatus = buildStatusRunning
		expected.Approvers = []string{}
		expected.Name = "my title"

		assert.Equal(t, expected, actual)
	})

	t.Run("get build status", func(t *testing.T) {
		mr := &gitlab.MergeRequest{}
		actual := gitlabFetcher.getPipelineStatus(mr)
		assert.Equal(t, buildStatusUnknown, actual)

		mr = &gitlab.MergeRequest{}
		mr.Pipeline = &gitlab.PipelineInfo{}
		actual = gitlabFetcher.getPipelineStatus(mr)
		assert.Equal(t, buildStatusUnknown, actual)

		mr = &gitlab.MergeRequest{}
		mr.Pipeline = &gitlab.PipelineInfo{}
		mr.Pipeline.Status = "failed"
		actual = gitlabFetcher.getPipelineStatus(mr)
		assert.Equal(t, buildStatusFailed, actual)

		mr = &gitlab.MergeRequest{}
		mr.Pipeline = &gitlab.PipelineInfo{}
		mr.Pipeline.Status = "success"
		actual = gitlabFetcher.getPipelineStatus(mr)
		assert.Equal(t, buildStatusSuccess, actual)
	})

	t.Run("get empty approvers", func(t *testing.T) {
		mr := &gitlab.MergeRequest{}
		mr.SourceProjectID = 12
		actual := gitlabFetcher.getApprovers(mr, 13)
		assert.Equal(t, []string{"foobar"}, actual)
	})

	t.Run("Render template with gitlabPullRequest()", func(t *testing.T) {
		tpl, err := util.CompileTemplate(`{{ $pr := gitlabPullRequest "12" "13" }}{{$pr.Name}} - {{$pr.Approvers}}`)
		assert.Nil(t, err)

		res, err := util.EvalTemplate(tpl, util.Parameters{})
		assert.Nil(t, err)
		assert.Equal(t, "Update XYZ - [foobar]", res)
	})

	t.Run("Render template with gitlabPullRequest() with not existing PR", func(t *testing.T) {
		tpl, err := util.CompileTemplate(`{{ gitlabPullRequest "12" "14" }}`)
		assert.Nil(t, err)

		_, err = util.EvalTemplate(tpl, util.Parameters{})
		assert.NotNil(t, err)
	})
}

func spawnTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// used in "get empty approvers" test
	mux.HandleFunc("/api/v4/projects/12/merge_requests/13/approvals", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(`{"id":1,"iid":1,"project_id":12,"title":"Update XYZ","state":"opened","created_at":"2021-07-09T16:13:24.990Z","updated_at":"2021-07-09T16:16:55.883Z","merge_status":"can_be_merged","approved":true,"approvals_required":0,"approvals_left":0,"require_password_to_approve":false,"approved_by":[{"user":{"id":1,"name":"FooBar","username":"foobar","state":"active"}}],"suggested_approvers":[],"approvers":[],"approver_groups":[],"user_has_approved":true,"user_can_approve":false,"approval_rules_left":[],"has_approval_rules":false,"merge_request_approvers_available":null,"multiple_approval_rules_available":null}`))
	})

	// returns an open PR
	mux.HandleFunc("/api/v4/projects/12/merge_requests/13", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(`{"id":1,"iid":1,"project_id":2,"title":"Update XYZ","description":"","state":"opened","created_at":"2021-07-09T16:13:24.990Z","updated_at":"2021-07-09T16:16:55.883Z","merged_by":null,"merged_at":null,"closed_by":null,"closed_at":null,"target_branch":"main","source_branch":"sdsd","user_notes_count":0,"upvotes":0,"downvotes":0,"author":{"id":1,"name":"foobar","username":"foobar","state":"active","avatar_url":"https://www.gravatar.com/avatar/e64c7d89f26bd1972efa854d13d7dd61?s=80\u0026d=identicon","web_url":"http://localhost/root"},"assignees":[],"assignee":null,"reviewers":[],"source_project_id":12,"target_project_id":2,"labels":[],"draft":false,"work_in_progress":false,"milestone":null,"merge_when_pipeline_succeeds":false,"merge_status":"can_be_merged","sha":"6ecdcb0f8bd667ee40777fd0ac02542505b7f722","merge_commit_sha":null,"squash_commit_sha":null,"discussion_locked":null,"should_remove_source_branch":null,"force_remove_source_branch":true,"reference":"!1","references":{"short":"!1","relative":"!1","full":"root/test!1"},"web_url":"http://localhost/root/test/-/merge_requests/1","time_stats":{"time_estimate":0,"total_time_spent":0,"human_time_estimate":null,"human_total_time_spent":null},"squash":false,"task_completion_status":{"count":0,"completed_count":0},"has_conflicts":false,"blocking_discussions_resolved":true,"approvals_before_merge":null,"subscribed":true,"changes_count":"1","latest_build_started_at":null,"latest_build_finished_at":null,"first_deployed_to_production_at":null,"pipeline":null,"head_pipeline":{"id":2,"project_id":2,"sha":"6ecdcb0f8bd667ee40777fd0ac02542505b7f722","ref":"sdsd","status":"pending","created_at":"2021-07-09T16:13:20.310Z","updated_at":"2021-07-09T16:13:20.626Z","web_url":"http://localhost/root/test/-/pipelines/2","before_sha":"0000000000000000000000000000000000000000","tag":false,"yaml_errors":null,"user":{"id":1,"name":"Administrator","username":"root","state":"active","avatar_url":"https://www.gravatar.com/avatar/e64c7d89f26bd1972efa854d13d7dd61?s=80\u0026d=identicon","web_url":"http://localhost/root"},"started_at":null,"finished_at":null,"committed_at":null,"duration":null,"queued_duration":null,"coverage":null,"detailed_status":{"icon":"status_pending","text":"pending","label":"pending","group":"pending","tooltip":"pending","has_details":true,"details_path":"/root/test/-/pipelines/2","illustration":null,"favicon":"/assets/ci_favicons/favicon_status_pending-5bdf338420e5221ca24353b6bff1c9367189588750632e9a871b7af09ff6a2ae.png"}},"diff_refs":{"base_sha":"eb86f422a87c05631959d11e2b5879a1d12d08b8","head_sha":"6ecdcb0f8bd667ee40777fd0ac02542505b7f722","start_sha":"eb86f422a87c05631959d11e2b5879a1d12d08b8"},"merge_error":null,"first_contribution":false,"user":{"can_merge":true}}`))
	})

	// not existing PR
	mux.HandleFunc("/api/v4/projects/12/merge_requests/14", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(`{"message":"404 Not found"`))
		res.WriteHeader(404)
	})

	return httptest.NewServer(mux)
}
