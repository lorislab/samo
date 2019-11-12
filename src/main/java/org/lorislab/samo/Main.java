package org.lorislab.samo;

import com.github.zafarkhaja.semver.Version;
import org.lorislab.samo.cli.CliUtil;
import org.lorislab.samo.maven.MavenProject;
import picocli.CommandLine.Command;
import picocli.CommandLine;

import java.io.File;

@Command(name = "samo",
        mixinStandardHelpOptions = true,
        version = "1.0",
        subcommands = {
                VersionCommand.class,
                CommandLine.HelpCommand.class, HelmCommand.class, DockerCommand.class, GitCommand.class
        }
)
public class Main extends CommonCommand implements Runnable {

    public static void main(String[] args) {
        int exitCode = new CommandLine(new Main()).execute(args);
        System.exit(exitCode);
    }

    public void run() {
        System.out.println("Main");
    }


    @CommandLine.Command( name = "git-version")
    public void gitVersion() {
        try {
            MavenProject project = MavenProject.loadFromFile(pom);
            System.out.println("Project: " + project.id);
            Version version = Version.valueOf(project.id.version.value);

            CliUtil.Return r = CliUtil.callCli("git rev-parse --short=" + length + " HEAD", "Error git sha", verbose);
            System.out.println("Git hash: " + r.response);
            version = version.setPreReleaseVersion(r.response);
            String releaseVersion = version.toString();
            project.setVersion(releaseVersion);
            System.out.println("Change version from " + project.id.version.value + " to " + releaseVersion);
        } catch (Exception ex) {
            ex.printStackTrace();
        }
    }

}
