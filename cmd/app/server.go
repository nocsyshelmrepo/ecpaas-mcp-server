package app

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"kubesphere.io/ks-mcp-server/cmd/app/options"
	"kubesphere.io/ks-mcp-server/pkg/k8s"
	"kubesphere.io/ks-mcp-server/pkg/kubesphere"
	"kubesphere.io/ks-mcp-server/pkg/tools/cluster"
	"kubesphere.io/ks-mcp-server/pkg/tools/userrole"
	"kubesphere.io/ks-mcp-server/pkg/tools/workspace"
	"kubesphere.io/ks-mcp-server/pkg/version"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "ks-mcp-server",
		Short:        "KubeSphere MCP Server",
		Version:      fmt.Sprint(version.Get()),
		SilenceUsage: true,
	}

	cmd.AddCommand(newVersionCommand())
	cmd.AddCommand(newStudioCommand())
	cmd.AddCommand(newSSECommand())

	return cmd
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of KubeSphere MCP Server",
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Println(version.Get())
		},
	}
}

func newStudioCommand() *cobra.Command {
	o := options.NewStdioOptions()
	cmd := &cobra.Command{
		Use:   "stdio",
		Short: "Start stdio server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// get restconfig
			restconfig, err := kubesphere.LoadKSConfig(o.KSConfig)
			if err != nil {
				return err
			}
			ksconfig, err := kubesphere.NewKSConfig(restconfig, o.KSApiserver)
			if err != nil {
				return err
			}

			k8sConfig, err := k8s.NewK8sClient()
			if err != nil {
				return err
			}

			// create MCP server
			s := server.NewMCPServer(
				"KubeSphere",
				"4.1.3",
			)
			s.SetTools(
				// register userrole
				userrole.ListUsers(ksconfig), userrole.GetUser(ksconfig), userrole.DeleteUser(ksconfig),
				userrole.ListRoles(ksconfig), userrole.GetRole(ksconfig), userrole.DeleteRole(ksconfig), userrole.ListPermissions(ksconfig),
				// register workspace
				workspace.ListWorkspaces(ksconfig), workspace.GetWorkspace(ksconfig), workspace.DeleteWorkspace(ksconfig), workspace.ListWorkspaceMembers(ksconfig), workspace.GetWorkspaceMember(ksconfig), workspace.DeleteWorkspaceMember(ksconfig), workspace.GetWorkspaceQuotas(ksconfig),
				workspace.ListApplicationRepos(ksconfig), workspace.ListApplications(ksconfig), workspace.GetApplication(ksconfig), workspace.DeleteApplication(ksconfig), workspace.GetApplicationVersion(ksconfig), workspace.ListProjectMembers(ksconfig),
				// register cluster
				cluster.ListClusters(ksconfig), cluster.GetCluster(ksconfig), cluster.DeleteCluster(ksconfig), cluster.ListClusterMembers(ksconfig), cluster.GetClustermember(ksconfig), cluster.DeleteClusterMember(ksconfig),
				cluster.ListNodes(ksconfig), cluster.GetNode(ksconfig), cluster.DeleteNode(ksconfig), cluster.ListProjects(ksconfig), cluster.GetProject(ksconfig), cluster.DeleteProject(ksconfig), cluster.ListDeployments(ksconfig), cluster.GetDeployment(ksconfig), cluster.CreateDeployment(ksconfig), cluster.DeleteDeployment(ksconfig), cluster.ScaleDeployment(ksconfig), cluster.ListStatefulsets(ksconfig), cluster.GetStatefulset(ksconfig), cluster.DeleteStatefulset(ksconfig), cluster.ScaleStatefulset(ksconfig), cluster.ListDaemonsets(ksconfig), cluster.GetDaemonset(ksconfig), cluster.DeleteDaemonset(ksconfig), cluster.ListJobs(ksconfig), cluster.GetJob(ksconfig), cluster.DeleteJob(ksconfig), cluster.ListCronJobs(ksconfig), cluster.GetCronjob(ksconfig), cluster.DeleteCronjob(ksconfig), cluster.ListPods(ksconfig), cluster.GetPod(ksconfig), cluster.LogsPod(ksconfig), cluster.DeletePod(ksconfig), cluster.ExecPod(k8sConfig), cluster.ListServices(ksconfig), cluster.GetService(ksconfig), cluster.DeleteService(ksconfig), cluster.ListIngresses(ksconfig), cluster.GetIngress(ksconfig), cluster.DeleteIngress(ksconfig), cluster.ListSecrets(ksconfig), cluster.GetSecret(ksconfig), cluster.DeleteSecret(ksconfig), cluster.ListConfigmaps(ksconfig), cluster.GetConfigmap(ksconfig), cluster.DeleteConfigmap(ksconfig), cluster.ListServiceAccounts(ksconfig), cluster.GetServiceaccount(ksconfig), cluster.DeleteServiceaccount(ksconfig), cluster.ListCustomResourceDefinitions(ksconfig), cluster.GetCustomResourceDefinition(ksconfig),
				cluster.ListPersistentVolumeClaims(ksconfig), cluster.GetPersistentvolumeclaim(ksconfig), cluster.DeletePersistentvolumeclaim(ksconfig), cluster.ListPersistentVolumes(ksconfig), cluster.GetPersistentvolume(ksconfig), cluster.DeletePersistentvolume(ksconfig), cluster.ListStorageClasses(ksconfig), cluster.GetStorageclass(ksconfig), cluster.DeleteStorageclass(ksconfig),
			)

			// Start the stdio server
			return server.ServeStdio(s)
		},
	}

	for _, f := range o.Flags().FlagSets {
		cmd.Flags().AddFlagSet(f)
	}
	return cmd
}

func newSSECommand() *cobra.Command {
	o := options.NewStdioOptions()
	cmd := &cobra.Command{
		Use:   "sse",
		Short: "Start sse server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// get restconfig
			restconfig, err := kubesphere.LoadKSConfig(o.KSConfig)
			if err != nil {
				return err
			}
			ksconfig, err := kubesphere.NewKSConfig(restconfig, o.KSApiserver)
			if err != nil {
				return err
			}

			k8sConfig, err := k8s.NewK8sClient()
			if err != nil {
				return err
			}

			// create MCP server
			s := server.NewMCPServer(
				"KubeSphere",
				"4.1.3",
			)
			s.SetTools(
				// register userrole
				userrole.ListUsers(ksconfig), userrole.GetUser(ksconfig), userrole.DeleteUser(ksconfig),
				userrole.ListRoles(ksconfig), userrole.GetRole(ksconfig), userrole.DeleteRole(ksconfig), userrole.ListPermissions(ksconfig),
				// register workspace
				workspace.ListWorkspaces(ksconfig), workspace.GetWorkspace(ksconfig), workspace.DeleteWorkspace(ksconfig), workspace.ListWorkspaceMembers(ksconfig), workspace.GetWorkspaceMember(ksconfig), workspace.DeleteWorkspaceMember(ksconfig), workspace.GetWorkspaceQuotas(ksconfig),
				workspace.ListApplicationRepos(ksconfig), workspace.ListApplications(ksconfig), workspace.GetApplication(ksconfig), workspace.DeleteApplication(ksconfig), workspace.GetApplicationVersion(ksconfig), workspace.ListProjectMembers(ksconfig),
				// register cluster
				cluster.ListClusters(ksconfig), cluster.GetCluster(ksconfig), cluster.DeleteCluster(ksconfig), cluster.ListClusterMembers(ksconfig), cluster.GetClustermember(ksconfig), cluster.DeleteClusterMember(ksconfig),
				cluster.ListNodes(ksconfig), cluster.GetNode(ksconfig), cluster.DeleteNode(ksconfig), cluster.ListProjects(ksconfig), cluster.GetProject(ksconfig), cluster.DeleteProject(ksconfig), cluster.ListDeployments(ksconfig), cluster.GetDeployment(ksconfig), cluster.CreateDeployment(ksconfig), cluster.DeleteDeployment(ksconfig), cluster.ScaleDeployment(ksconfig), cluster.ListStatefulsets(ksconfig), cluster.GetStatefulset(ksconfig), cluster.DeleteStatefulset(ksconfig), cluster.ScaleStatefulset(ksconfig), cluster.ListDaemonsets(ksconfig), cluster.GetDaemonset(ksconfig), cluster.DeleteDaemonset(ksconfig), cluster.ListJobs(ksconfig), cluster.GetJob(ksconfig), cluster.DeleteJob(ksconfig), cluster.ListCronJobs(ksconfig), cluster.GetCronjob(ksconfig), cluster.DeleteCronjob(ksconfig), cluster.ListPods(ksconfig), cluster.GetPod(ksconfig), cluster.LogsPod(ksconfig), cluster.DeletePod(ksconfig), cluster.ExecPod(k8sConfig), cluster.ListServices(ksconfig), cluster.GetService(ksconfig), cluster.DeleteService(ksconfig), cluster.ListIngresses(ksconfig), cluster.GetIngress(ksconfig), cluster.DeleteIngress(ksconfig), cluster.ListSecrets(ksconfig), cluster.GetSecret(ksconfig), cluster.DeleteSecret(ksconfig), cluster.ListConfigmaps(ksconfig), cluster.GetConfigmap(ksconfig), cluster.DeleteConfigmap(ksconfig), cluster.ListServiceAccounts(ksconfig), cluster.GetServiceaccount(ksconfig), cluster.DeleteServiceaccount(ksconfig), cluster.ListCustomResourceDefinitions(ksconfig), cluster.GetCustomResourceDefinition(ksconfig),
				cluster.ListPersistentVolumeClaims(ksconfig), cluster.GetPersistentvolumeclaim(ksconfig), cluster.DeletePersistentvolumeclaim(ksconfig), cluster.ListPersistentVolumes(ksconfig), cluster.GetPersistentvolume(ksconfig), cluster.DeletePersistentvolume(ksconfig), cluster.ListStorageClasses(ksconfig), cluster.GetStorageclass(ksconfig), cluster.DeleteStorageclass(ksconfig),
			)
			// Start the sse server
			sse := server.NewSSEServer(s)
			return sse.Start(":8080")
		},
	}

	for _, f := range o.Flags().FlagSets {
		cmd.Flags().AddFlagSet(f)
	}
	return cmd
}
