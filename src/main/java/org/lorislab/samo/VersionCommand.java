package org.lorislab.samo;

import com.github.zafarkhaja.semver.Version;
import org.lorislab.samo.cli.CliUtil;
import org.lorislab.samo.maven.MavenProject;
import picocli.CommandLine;

import java.nio.file.Path;
import java.util.Optional;

@CommandLine.Command(name = "version", subcommands = {
        VersionCommand.Release.class,
        VersionCommand.Git.class
})
public class VersionCommand extends CommonCommand {

    @Override
    public void run() {
        try {
            MavenProject project = MavenProject.loadFromFile(pom);
            System.out.println(project.id.version.value);
        } catch (Exception ex) {
            ex.printStackTrace();
        }
    }

    @CommandLine.Command(name = "release")
    public static class Release extends CommonCommand {

        @Override
        public void run() {
            try {
                MavenProject project = MavenProject.loadFromFile(pom);
                System.out.println("Project: " + project.id);
                Version version = Version.valueOf(project.id.version.value);
                String releaseVersion = version.getNormalVersion();
                project.setVersion(releaseVersion);
                System.out.println("Change version from " + project.id.version.value + " to " + releaseVersion);
            } catch (Exception ex) {
                ex.printStackTrace();
            }
        }
    }

    @CommandLine.Command(name = "git")
    public static class Git extends CommonCommand {

        @Override
        public void run() {
            try {
                MavenProject project = MavenProject.loadFromFile(pom);
                if (verbose) {
                    System.out.println("Project: " + project.id);
                }
                Version version = Version.valueOf(project.id.version.value);

                CliUtil.Return r = CliUtil.callCli(getGitPath() + " rev-parse --short=" + length + " HEAD", "Error git sha", verbose);
                if (verbose) {
                    System.out.println("Git hash: " + r.response);
                }
                version = version.setPreReleaseVersion(r.response);
                String releaseVersion = version.toString();
                project.setVersion(releaseVersion);
                System.out.println("Change version from " + project.id.version.value + " to " + releaseVersion + " in the file: " + project.file);
            } catch (Exception ex) {
                ex.printStackTrace();
            }
        }
    }

    static Path getGitPath() {
        String helmExecutable = CliUtil.IS_WINDOWS ? "git.exe" : "git";
        Optional<Path> path = CliUtil.findInPath(helmExecutable);
        return path.orElseThrow(() -> new RuntimeException("git executable is not found."));
    }

}
