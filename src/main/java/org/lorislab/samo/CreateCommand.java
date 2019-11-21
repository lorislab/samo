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
 * The create command.
 */
@CommandLine.Command(name = "create",
        description = "Create project commands",
        mixinStandardHelpOptions = true,
        showDefaultValues = true,
        subcommands = {
                CreateCommand.Release.class,
                CreateCommand.Patch.class
        }
)
class CreateCommand extends SamoCommand {

    /**
     * Create a release tag of the project and increase development version.
     */
    @CommandLine.Command(name = "release",
            mixinStandardHelpOptions = true,
            showDefaultValues = true,
            description = "Create release of the current project and state")
    public static class Release extends SamoCommand {

        /**
         * The maven options.
         */
        @CommandLine.Mixin
        MavenOptions maven;

        /**
         * The message for the commit.
         */
        @CommandLine.Option(
                names = {"-m", "--message"},
                defaultValue = "Development version ",
                required = true,
                description = "commit message for new development version"
        )
        String message;

        /**
         * Create a release tag of the project and increase development version.
         *
         * @return the exit code.
         * @throws Exception if the method fails.
         */
        @Override
        public Integer call() throws Exception {
            MavenProject project = getMavenProject(maven.pom);
            Version version = Version.valueOf(project.id.version.value);
            String releaseVersion = version.getNormalVersion();

            try {
                // git create tag
                cmd("git tag " + releaseVersion, "Error create tag");

                // maven change version
                Version newVersion;
                if (version.getPatchVersion() == 0) {
                    newVersion = version.incrementMinorVersion(SNAPSHOT);
                } else {
                    newVersion = version.incrementPatchVersion(SNAPSHOT);
                }
                project.setVersion(newVersion.toString());
                info("Change version from %s to %s in the file: %s", project.id.version.value, newVersion, project.file);

                // git commit & push
                cmd("git add .", "Error git add");
                cmd("git commit -m \"" + message + newVersion + "\"", "Error git commit");
                cmd("git push origin refs/heads/*:refs/heads/* refs/tags/*:refs/tags/*", "Error git push");

            } catch (Exception ex) {
                cmd("rm -f .git/index.lock", "Error remove git index");
                throw ex;
            }
            return CommandLine.ExitCode.OK;
        }
    }

    @CommandLine.Command(name = "patch",
            mixinStandardHelpOptions = true,
            showDefaultValues = true,
            description = "Create patch of the release")
    public static class Patch extends SamoCommand {

        /**
         * The maven options.
         */
        @CommandLine.Mixin
        MavenOptions maven;

        /**
         * The message for the commit.
         */
        @CommandLine.Option(
                names = {"-m", "--message"},
                defaultValue = "Create patch version ",
                required = true,
                description = "commit message for patch version"
        )
        String message;

        /**
         * The release tag
         */
        @CommandLine.Option(
                names = {"-t", "--tag"},
                required = true,
                interactive = true,
                description = "The release version (tag x.x.0) to patch"
        )
        String tag;

        /**
         * Create a patch branch for the release tag of the project and increase patch development version.
         *
         * @return the exit code.
         * @throws Exception if the method fails.
         */
        @Override
        public Integer call() throws Exception {

            Version version = Version.valueOf(tag);
            Version releaseVersion = Version.valueOf(version.getNormalVersion());
            if (!version.equals(releaseVersion)) {
                throw new RuntimeException("The release version does not have patch 0 or empty prerelease suffix.");
            }

            Version pv = releaseVersion.incrementPatchVersion(SNAPSHOT);

            try {
                // create & checkout branch
                String branchName = releaseVersion.getMajorVersion() + "." + releaseVersion.getMinorVersion();
                cmd("git checkout -b " + branchName + " " + version, "Error create and checkout branch");

                // change version
                MavenProject project = getMavenProject(maven.pom);
                project.setVersion(pv.toString());

                // git commit & push
                cmd("git add .", "Error git add");
                cmd("git commit -m \"" + message + pv + "\"", "Error git commit");
                cmd("git push -u origin " + branchName, "Error git push branch");
            } catch (Exception ex) {
                cmd("rm -f .git/index.lock", "Error remove git index");
                throw ex;
            }
            return CommandLine.ExitCode.OK;
        }
    }

}
