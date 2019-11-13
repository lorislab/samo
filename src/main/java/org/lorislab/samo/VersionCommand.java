package org.lorislab.samo;

import com.github.zafarkhaja.semver.Version;
import org.lorislab.samo.data.MavenProject;
import picocli.CommandLine;

@CommandLine.Command(name = "version",
        description = "All version commands",
        subcommands = {
                VersionCommand.Release.class,
                VersionCommand.Git.class,
                VersionCommand.Info.class,
                VersionCommand.Snapshot.class

        }
)
public class VersionCommand extends CommonCommand {

    @CommandLine.Spec
    CommandLine.Model.CommandSpec spec;

    @Override
    public Integer call() {
        spec.commandLine().usage(System.out);
        return CommandLine.ExitCode.OK;
    }

    @CommandLine.Command(name = "info", description = "Show current version")
    public static class Info extends CommonCommand {

        @Override
        public Integer call() throws Exception {
            MavenProject project = MavenProject.loadFromFile(pom);
            logInfo(project.id.version.value);
            return CommandLine.ExitCode.OK;
        }
    }

    @CommandLine.Command(name = "release", description = "Set the release version")
    public static class Release extends CommonCommand {

        @Override
        public Integer call() throws Exception {
            MavenProject project = MavenProject.loadFromFile(pom);
            logVerbose("Project: " + project.id);
            Version version = Version.valueOf(project.id.version.value);
            String releaseVersion = version.getNormalVersion();
            logVerbose(releaseVersion);
            project.setVersion(releaseVersion);
            logInfo("Change version from " + project.id.version.value + " to " + releaseVersion);
            return CommandLine.ExitCode.OK;
        }
    }

    @CommandLine.Command(name = "snapshot", description = "Set snapshot prerelease version")
    public static class Snapshot extends CommonCommand {

        @Override
        public Integer call() throws Exception {
            MavenProject project = MavenProject.loadFromFile(pom);
            logVerbose("Project: " + project.id);
            Version version = Version.valueOf(project.id.version.value);
            version = version.setPreReleaseVersion("SNAPSHOT");
            String releaseVersion = version.toString();
            logVerbose(releaseVersion);

            project.setVersion(releaseVersion);
            logInfo("Change version from " + project.id.version.value + " to " + releaseVersion + " in the file: " + project.file);
            return CommandLine.ExitCode.OK;
        }
    }

    @CommandLine.Command(name = "sha-prerelease", description = "Set git sha prerelease version")
    public static class Git extends CommonCommand {

        @Override
        public Integer call() throws Exception {
            MavenProject project = MavenProject.loadFromFile(pom);
            logVerbose("Project: " + project.id);

            Version version = Version.valueOf(project.id.version.value);

            Return r = callCli(getGitPath() + " rev-parse --short=" + length + " HEAD", "Error git sha", verbose);
            logVerbose("Git hash: " + r.response);

            version = version.setPreReleaseVersion(r.response);
            String releaseVersion = version.toString();
            logVerbose(releaseVersion);

            project.setVersion(releaseVersion);
            logInfo("Change version from " + project.id.version.value + " to " + releaseVersion + " in the file: " + project.file);
            return CommandLine.ExitCode.OK;
        }
    }


}
