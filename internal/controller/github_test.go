package controller_test

import (
	"context"

	"github.com/google/go-github/v33/github"
	harmonizeriov1 "github.com/ibexmonj/harmonizer/api/v1beta1"
	"github.com/ibexmonj/harmonizer/internal/controller"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("GitHub", func() {
	var (
		ctx        context.Context
		fakeClient *controller.TeamReconciler
		ghClient   *MockGitHubClient
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme, _ := harmonizeriov1.SchemeBuilder.Build()
		fakeClient = &controller.TeamReconciler{
			Client: fake.NewClientBuilder().WithScheme(scheme).Build(),
			Scheme: scheme,
		}
		ghClient = &MockGitHubClient{}
	})

	FDescribe("FetchTeams", func() {
		It("should fetch teams from GitHub", func() {
			teams, err := controller.FetchTeams(ctx, ghClient)
			Expect(err).NotTo(HaveOccurred())
			Expect(teams).To(HaveLen(1))
			Expect(*teams[0].Name).To(Equal("team1"))
		})
	})

	Describe("CreateTeamResource", func() {
		It("should create a new Team resource in Kubernetes if it doesn't already exist", func() {
			team := &github.Team{Name: github.String("team1")}
			err := controller.CreateTeamResource(ctx, fakeClient, team, ghClient)
			Expect(err).NotTo(HaveOccurred())

			var createdTeam harmonizeriov1.Team
			err = fakeClient.Get(ctx, client.ObjectKey{Name: "team1", Namespace: "default"}, &createdTeam)
			Expect(err).NotTo(HaveOccurred())
			Expect(createdTeam.Name).To(Equal("team1"))
		})
	})

	Describe("FetchAndCreateTeams", func() {
		It("should fetch teams from GitHub and create corresponding resources in Kubernetes", func() {
			err := controller.FetchAndCreateTeams(ctx, fakeClient, ctrl.Request{}, ghClient)
			Expect(err).NotTo(HaveOccurred())

			var createdTeam harmonizeriov1.Team
			err = fakeClient.Get(ctx, client.ObjectKey{Name: "team1", Namespace: "default"}, &createdTeam)
			Expect(err).NotTo(HaveOccurred())
			Expect(createdTeam.Name).To(Equal("team1"))
		})
	})
})

type MockGitHubClient struct{}

func (f *MockGitHubClient) ListTeams(ctx context.Context, org string, opt *github.ListOptions) ([]*github.Team, *github.Response, error) {
	return []*github.Team{{Name: github.String("team1")}}, &github.Response{}, nil
}

func (f *MockGitHubClient) ListTeamMembersBySlug(ctx context.Context, org string, slug string, opt *github.TeamListTeamMembersOptions) ([]*github.User, *github.Response, error) {
	return []*github.User{{Login: github.String("user1")}}, &github.Response{}, nil
}
