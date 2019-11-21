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

import com.github.zafarkhaja.semver.Version;
import org.lorislab.samo.data.MavenProject;
import picocli.CommandLine;

/**
 * The maven command.
 */
@CommandLine.Command(name = "maven",
        description = "Maven version commands",
        mixinStandardHelpOptions = true,
        subcommands = {
                MavenCommand.MavenVersion.class,
                MavenCommand.Release.class,
                MavenCommand.Git.class,
                MavenCommand.Snapshot.class
        }
)
class MavenCommand extends SamoCommand {

    /**
     * The maven version command.
     */
    @CommandLine.Command(name = "version",
            mixinStandardHelpOptions = true,
            description = "Show current maven version")
    public static class MavenVersion extends SamoCommand {

        /**
         * The maven options.
         */
        @CommandLine.Mixin
        MavenOptions maven;

        /**
         * Returns the current version of the maven project.
         *
         * @return the exit code.
         * @throws Exception if the method fails.
         */
        @Override
        public Integer call() throws Exception {
            MavenProject project = getMavenProject(maven.pom);
            output(project.id.version.value);
            return CommandLine.ExitCode.OK;
        }
    }

    /**
     * Sets the maven project to release version.
     */
    @CommandLine.Command(name = "set-release",
            mixinStandardHelpOptions = true,
            description = "Set the release version")
    public static class Release extends SamoCommand {

        /**
         * The maven options.
         */
        @CommandLine.Mixin
        MavenOptions maven;

        /**
         * Sets the maven project to release version.
         *
         * @return the exit code.
         * @throws Exception if the method fails.
         */
        @Override
        public Integer call() throws Exception {
            MavenProject project = getMavenProject(maven.pom);
            Version version = Version.valueOf(project.id.version.value);
            setMavenVersion(project, version.getNormalVersion());
            return CommandLine.ExitCode.OK;
        }
    }

    /**
     * Sets the maven project to snapshot version.
     */
    @CommandLine.Command(name = "set-snapshot",
            mixinStandardHelpOptions = true,
            description = "Set snapshot prerelease version")
    public static class Snapshot extends SamoCommand {

        /**
         * The maven options.
         */
        @CommandLine.Mixin
        MavenOptions maven;

        /**
         * Sets the maven project to snapshot version.
         *
         * @return the exit code.
         * @throws Exception if the method fails.
         */
        @Override
        public Integer call() throws Exception {
            MavenProject project = getMavenProject(maven.pom);
            String releaseVersion = preReleaseVersion(project, SNAPSHOT);
            setMavenVersion(project, releaseVersion);
            return CommandLine.ExitCode.OK;
        }
    }

    /**
     * Sets the maven project to git hash prerelease version.
     */
    @CommandLine.Command(name = "set-hash",
            mixinStandardHelpOptions = true,
            description = "Set git hash prerelease version")
    public static class Git extends SamoCommand {

        /**
         * The maven options.
         */
        @CommandLine.Mixin
        MavenOptions maven;

        /**
         * The git options.
         */
        @CommandLine.Mixin
        GitOptions git;

        /**
         * Sets the maven project to git sha prerelease version.
         *
         * @return the exit code.
         * @throws Exception if the method fails.
         */
        @Override
        public Integer call() throws Exception {
            MavenProject project = getMavenProject(maven.pom);
            String releaseVersion = preReleaseVersion(project, gitHash(git));
            setMavenVersion(project, releaseVersion);
            return CommandLine.ExitCode.OK;
        }
    }

}
