/*
 * Copyright 2019 lorislab.org.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package org.lorislab.samo;

import org.lorislab.samo.data.MavenProject;
import picocli.CommandLine;

import java.util.concurrent.Callable;

/**
 * The maven command.
 */
@CommandLine.Command(name = "maven",
        description = "Maven version commands",
        subcommands = {
                MavenCommand.Version.class,
                MavenCommand.Release.class,
                MavenCommand.Git.class,
                MavenCommand.Snapshot.class
        }
)
public class MavenCommand implements Callable<Integer> {

    /**
     * The command specification.
     */
    @CommandLine.Spec
    CommandLine.Model.CommandSpec spec;

    /**
     * Show help of the maven commands.
     *
     * @return the exit code.
     */
    @Override
    public Integer call() {
        spec.commandLine().usage(System.out);
        return CommandLine.ExitCode.OK;
    }

    /**
     * The maven version command.
     */
    @CommandLine.Command(name = "version", description = "Show current maven version")
    public static class Version extends CommonCommand {

        /**
         * Returns the current version of the maven project.
         *
         * @return the exit code.
         * @throws Exception if the method fails.
         */
        @Override
        public Integer call() throws Exception {
            MavenProject project = getMavenProject();
            logInfo(project.id.version.value);
            return CommandLine.ExitCode.OK;
        }
    }

    /**
     * Sets the maven project to release version.
     */
    @CommandLine.Command(name = "set-release", description = "Set the release version")
    public static class Release extends CommonCommand {

        /**
         * Sets the maven project to release version.
         *
         * @return the exit code.
         * @throws Exception if the method fails.
         */
        @Override
        public Integer call() throws Exception {
            MavenProject project = getMavenProject();
            com.github.zafarkhaja.semver.Version version = com.github.zafarkhaja.semver.Version.valueOf(project.id.version.value);
            String releaseVersion = version.getNormalVersion();
            logVerbose(releaseVersion);
            project.setVersion(releaseVersion);
            logInfo("Change version from " + project.id.version.value + " to " + releaseVersion + " in the file: " + project.file);
            return CommandLine.ExitCode.OK;
        }
    }

    /**
     * Sets the maven project to snapshot version.
     */
    @CommandLine.Command(name = "set-snapshot", description = "Set snapshot prerelease version")
    public static class Snapshot extends CommonCommand {

        /**
         * Sets the maven project to snapshot version.
         *
         * @return the exit code.
         * @throws Exception if the method fails.
         */
        @Override
        public Integer call() throws Exception {
            MavenProject project = getMavenProject();
            com.github.zafarkhaja.semver.Version version = com.github.zafarkhaja.semver.Version.valueOf(project.id.version.value);
            version = version.setPreReleaseVersion(SNAPSHOT);
            String releaseVersion = version.toString();
            logVerbose(releaseVersion);

            project.setVersion(releaseVersion);
            logInfo("Change version from " + project.id.version.value + " to " + releaseVersion + " in the file: " + project.file);
            return CommandLine.ExitCode.OK;
        }
    }

    /**
     * Sets the maven project to git hash prerelease version.
     */
    @CommandLine.Command(name = "set-hash", description = "Set git hash prerelease version")
    public static class Git extends CommonCommand {

        /**
         * The length of the git sha
         */
        @CommandLine.Option(
                names = {"-l", "--length"},
                paramLabel = "LENGTH",
                defaultValue = "7",
                required = true,
                description = "the git sha length"
        )
        int length;

        /**
         * Sets the maven project to git sha prerelease version.
         *
         * @return the exit code.
         * @throws Exception if the method fails.
         */
        @Override
        public Integer call() throws Exception {
            MavenProject project = getMavenProject();
            com.github.zafarkhaja.semver.Version version = com.github.zafarkhaja.semver.Version.valueOf(project.id.version.value);

            Return r = cmd("git rev-parse --short=" + length + " HEAD", "Error git sha");
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
