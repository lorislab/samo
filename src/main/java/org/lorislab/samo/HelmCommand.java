package org.lorislab.samo;

import com.github.zafarkhaja.semver.Version;
import org.lorislab.samo.data.MavenProject;
import org.lorislab.samo.data.YamlNode;
import picocli.CommandLine;

import java.nio.file.Path;
import java.nio.file.Paths;

/**
 * The helm command
 */
@CommandLine.Command(name = "helm",
        description = "Helm commands",
        mixinStandardHelpOptions = true,
        showDefaultValues = true,
        usageHelpAutoWidth = true,
        subcommands = {
                HelmCommand.AddRepo.class,
                HelmCommand.Update.class,
                HelmCommand.Build.class,
                HelmCommand.Push.class,
                HelmCommand.Release.class
        }
)
class HelmCommand extends SamoCommand {

    /**
     * Add repo helm command.
     */
    @CommandLine.Command(name = "add-repo",
            mixinStandardHelpOptions = true,
            usageHelpAutoWidth = true,
            showDefaultValues = true,
            description = "Add the helm chart repository.")
    public static class AddRepo extends HelmSubCommand {

        /**
         * The helm chart user.
         */
        @CommandLine.Option(
                names = {"-u", "--username"},
                defaultValue = "${env:SAMO_HELM_USERNAME}",
                description = "the helm chart repository username"
        )
        String username;

        /**
         * The helm chart user.
         */
        @CommandLine.Option(
                names = {"-p", "--password"},
                defaultValue = "${env:SAMO_HELM_PASSWORD}",
                description = "the helm chart repository password"
        )
        String password;

        /**
         * The helm chart repository name.
         */
        @CommandLine.Option(
                names = {"-r", "--repo-name"},
                defaultValue = "${env:SAMO_HELM_NAME}",
                required = true,
                description = "the helm chart repository name"
        )
        String name;

        /**
         * The helm chart repository URL.
         */
        @CommandLine.Option(
                names = {"-e", "--repo-url"},
                defaultValue = "${env:SAMO_HELM_URL}",
                required = true,
                description = "the helm chart repository name"
        )
        String url;

        /**
         * {@inheritDoc }
         */
        @Override
        public Integer call() {
            StringBuilder sb = new StringBuilder();
            sb.append("helm repo add ").append(url);
            if (username != null && password != null) {
                sb.append(" --password ").append(password);
                sb.append(" --username ").append(username);
            }
            cmd(sb.toString(), "Error add helm chart repository");
            info("Add the helm chart repository %s - %s", name, url);
            return CommandLine.ExitCode.OK;
        }

    }

    /**
     * Update repo helm command.
     */
    @CommandLine.Command(name = "update",
            mixinStandardHelpOptions = true,
            showDefaultValues = true,
            description = "Update helm chart repositories for the project.")
    public static class Update extends HelmSubCommand {

        /**
         * {@inheritDoc }
         */
        @Override
        public Integer call() throws Exception {
            updateHelmRepositories();
            info("Helm chart repositories updated.");
            return CommandLine.ExitCode.OK;
        }

    }

    /**
     * Build helm command.
     */
    @CommandLine.Command(name = "build",
            mixinStandardHelpOptions = true,
            showDefaultValues = true,
            description = "Build helm chart for the project.")
    public static class Build extends HelmSubCommand {

        /**
         * The helm options.
         */
        @CommandLine.Mixin
        HelmOptions helm;

        /**
         * The maven options.
         */
        @CommandLine.Mixin
        MavenOptions maven;

        /**
         * {@inheritDoc }
         */
        @Override
        public Integer call() throws Exception {
            MavenProject project = getMavenProject(maven.pom);

            // update the helm repository
            updateHelmRepositories();

            // package helm chart
            packageHelmChart(project, helm, project.id.version.value);
            return CommandLine.ExitCode.OK;
        }

    }

    /**
     * Push helm command.
     */
    @CommandLine.Command(name = "push",
            mixinStandardHelpOptions = true,
            showDefaultValues = true,
            description = "Push helm chart for the project.")
    public static class Push extends HelmSubCommand {

        /**
         * The maven options.
         */
        @CommandLine.Mixin
        MavenOptions maven;

        /**
         * The helm repo options.
         */
        @CommandLine.Mixin
        HelmRepoOptions repo;

        /**
         * {@inheritDoc }
         */
        @Override
        public Integer call() throws Exception {

            MavenProject project = getMavenProject(maven.pom);

            // push helm chart
            pushHelmChart(project, repo);
            return CommandLine.ExitCode.OK;
        }
    }

    /**
     * Release helm command.
     */
    @CommandLine.Command(name = "release",
            mixinStandardHelpOptions = true,
            showDefaultValues = true,
            description = "Release helm chart for the project.")
    public static class Release extends HelmSubCommand {

        /**
         * The helm options.
         */
        @CommandLine.Mixin
        HelmOptions helm;

        /**
         * The git options.
         */
        @CommandLine.Mixin
        GitOptions git;

        /**
         * The maven options.
         */
        @CommandLine.Mixin
        MavenOptions maven;

        /**
         * The helm repo options.
         */
        @CommandLine.Mixin
        HelmRepoOptions repo;

        /**
         * {@inheritDoc }
         */
        @Override
        public Integer call() throws Exception {
            // update the helm repository
            updateHelmRepositories();

            MavenProject project = getMavenProject(maven.pom);
            Version version = Version.valueOf(project.id.version.value).setPreReleaseVersion(gitHash(git));

            // helm pull
            cmd("helm pull " + repo + "/" + project.id.artifactId.value + " --version " + version + " --untar --untardir " + helm.dir, "Error update helm repositories");

            // helm package
            packageHelmChart(project, helm, version.getNormalVersion());

            // push helm chart
            pushHelmChart(project, repo);

            return CommandLine.ExitCode.OK;
        }
    }

    /**
     * Helm sub-command.
     */
    static class HelmSubCommand extends SamoCommand {

        /**
         * Update the helm repositories.
         */
        void updateHelmRepositories() {
            cmd("helm repo update", "Error update helm repositories");
        }

        /**
         * Package the helm chart
         *
         * @param project the maven project.
         * @param helm    the helm options.
         * @param version the version.
         */
        void packageHelmChart(MavenProject project, HelmOptions helm, String version) {
            Path path = Paths.get(helm.dir).resolve(project.id.artifactId.value);
            cmd("helm package " + path + " -u --app-version " + version + " --version " + version, "Error build helm chart");
            info("helm package %s", path);
        }

        /**
         * Gets the helm chart URL by name.
         * @param name the helm chart name.
         * @return the corresponding URL.
         */
        String getHelmChartRepositoryUrl(String name) {
            String url = null;
            Return r = cmd("helm repo list -o yaml", "Error get list of helm chart repositories");
            debug("repositories:\n%s", r.response);
            YamlNode node = YamlNode.parse(r.response);
            if (node != null && !node.isEmpty()) {
                for (YamlNode n : node) {
                    if (name.equals(n.get("name").value)) {
                        url = n.get("url").value;
                    }
                }
            }
            info("Repo %s url %s", name, url);
            return url;
        }

        /**
         * Push the helm chart
         *
         * @param project the maven project.
         * @param repo    the helm chart repository options.
         * @throws Exception if the method fails.
         */
        void pushHelmChart(MavenProject project, HelmRepoOptions repo) throws Exception {
            // get the helm repository URL
            if (repo.url == null || repo.url.isEmpty()) {
                repo.url = getHelmChartRepositoryUrl(repo.name);
            }
            if (repo.url == null || repo.url.isEmpty()) {
                throw new RuntimeException("The helm chart repository URL is not defined!");
            }

            // create helm push command
            StringBuilder sb = new StringBuilder();
            sb.append("curl -is ");
            if (repo.username != null && repo.password != null) {
                sb.append(" -u ").append(repo.username).append(":").append(repo.password);
            }

            sb.append(" ").append(repo.url).append(" --upload-file ").append(project.id.artifactId.value).append(".tgz");
            cmd(sb.toString(), "Error push helm chart");
        }

    }

    /**
     * The helm chart repository options.
     */
    static class HelmRepoOptions {

        /**
         * The helm chart user.
         */
        @CommandLine.Option(
                names = {"-u", "--username"},
                defaultValue = "${env:SAMO_HELM_USERNAME}",
                description = "the helm chart repository username"
        )
        String username;

        /**
         * The helm chart user.
         */
        @CommandLine.Option(
                names = {"-p", "--password"},
                defaultValue = "${env:SAMO_HELM_PASSWORD}",
                description = "the helm chart repository password"
        )
        String password;

        /**
         * The helm chart repository name.
         */
        @CommandLine.Option(
                names = {"-r", "--repo-name"},
                defaultValue = "${env:SAMO_HELM_NAME}",
                description = "the helm chart repository name"
        )
        String name;

        /**
         * The helm chart repository URL.
         */
        @CommandLine.Option(
                names = {"-e", "--repo-url"},
                defaultValue = "${env:SAMO_HELM_URL}",
                description = "the helm chart repository name"
        )
        String url;
    }

    /**
     * The helm options.
     */
    public static class HelmOptions {

        /**
         * The helm chart directory.
         */
        @CommandLine.Option(
                names = {"-d", "--chart-dir"},
                required = true,
                defaultValue = "target/helm",
                description = "the helm chart directory"
        )
        String dir;

    }
}
