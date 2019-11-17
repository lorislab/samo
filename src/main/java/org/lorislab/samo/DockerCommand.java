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
import com.github.zafarkhaja.semver.Version;

import java.util.concurrent.Callable;

/**
 * The maven command.
 */
@CommandLine.Command(name = "docker",
        description = "Docker version commands",
        subcommands = {
                DockerCommand.Release.class,
                DockerCommand.Build.class,
                DockerCommand.Push.class
        }
)
public class DockerCommand implements Callable<Integer> {

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
     * Sets the maven project to release version.
     */
    @CommandLine.Command(name = "release", description = "Set the release version")
    public static class Release extends CommonCommand {

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
         * The docker repository
         */
        @CommandLine.Option(
                names = {"-i", "--image"},
                paramLabel = "IMAGE",
                description = "the docker image. Default value maven project artifactId."
        )
        String image;

        /**
         * The docker repository
         */
        @CommandLine.Option(
                names = {"-r", "--repository"},
                paramLabel = "REPOSITORY",
                defaultValue = "docker.io",
                required = true,
                description = "the docker repository"
        )
        String repository;

        /**
         * The docker repository
         */
        @CommandLine.Option(
                names = {"-rr", "--release-repository"},
                paramLabel = "REPOSITORY",
                description = "the docker release repository"
        )
        String releaseRepository;

        /**
         * Sets the maven project to release version.
         *
         * @return the exit code.
         * @throws Exception if the method fails.
         */
        @Override
        public Integer call() throws Exception {
            MavenProject project = getMavenProject();

            Return r = cmd("git rev-parse --short=" + length + " HEAD", "Error git sha");
            logVerbose("Git hash: " + r.response);

            Version version = Version.valueOf(project.id.version.value);
            Version pullVersion = version.setPreReleaseVersion(r.response);

            // set the docker release repository
            if (releaseRepository == null || releaseRepository.isEmpty()) {
                releaseRepository = repository;
            }

            // set the docker image name.
            if (image == null || image.isEmpty()) {
                image = project.id.artifactId.value;
            }
            String imageRelease = releaseRepository + "/" + image + ":" + version.getNormalVersion();
            String imagePull = repository + "/" + image + ":" + pullVersion;

            // execute the docker commands
            cmd("docker pull " + imagePull, "Error pull docker image");
            cmd("docker tag " + imagePull + " " + imageRelease, "Error tag docker image");
            cmd("docker push " + imageRelease, "Error docker push image");

            logInfo("Docker push new release image: " + imageRelease);
            return CommandLine.ExitCode.OK;
        }
    }

    /**
     * Sets the maven project to release version.
     */
    @CommandLine.Command(name = "build", description = "Build docker version")
    public static class Build extends CommonCommand {

        /**
         * The docker repository
         */
        @CommandLine.Option(
                names = {"-i", "--image"},
                paramLabel = "IMAGE",
                description = "the docker image. Default value maven project artifactId."
        )
        String image;

        /**
         * The docker repository
         */
        @CommandLine.Option(
                names = {"-r", "--repository"},
                paramLabel = "REPOSITORY",
                defaultValue = "docker.io",
                required = true,
                description = "the docker repository"
        )
        String repository;

        /**
         * The docker repository
         */
        @CommandLine.Option(
                names = {"-f", "--dockerfile"},
                paramLabel = "DOCKERFILE",
                defaultValue = ".",
                required = true,
                description = "the docker file"
        )
        String dockerfile;


        /**
         * Verbose flag.
         */
        @CommandLine.Option(
                names = {"-v", "--branch"},
                defaultValue = "true",
                required = true,
                description = "tag the docker image with a branch name"
        )
        boolean branch;

        /**
         * Sets the maven project to release version.
         *
         * @return the exit code.
         * @throws Exception if the method fails.
         */
        @Override
        public Integer call() throws Exception {
            MavenProject project = getMavenProject();

            // set the docker image name.
            if (image == null || image.isEmpty()) {
                image = project.id.artifactId.value;
            }

            StringBuilder log = new StringBuilder();
            log.append("Docker build new images [");

            StringBuilder sb = new StringBuilder();
            sb.append("docker build");

            String imageName = repository + "/" + image + ":" + project.id.version.value;
            sb.append(" -t ").append(imageName);
            log.append(imageName);

            String branchTag = "";
            if (branch) {
                Return r = cmd("git rev-parse --abbrev-ref HEAD", "Error git branch name");
                logVerbose("Git branch: " + r.response);
                branchTag = repository + "/" + image + ":" + r.response;
                sb.append(" -t ").append(branchTag);
                log.append(",").append(branchTag);
            }
            sb.append(" ").append(dockerfile);
            log.append("]");

            // execute the docker commands
            cmd(sb.toString(), "Error build docker image");

            logInfo(log.toString());
            return CommandLine.ExitCode.OK;
        }
    }

    /**
     * Docker push command
     */
    @CommandLine.Command(name = "push", description = "Push docker version")
    public static class Push extends CommonCommand {

        /**
         * The docker repository
         */
        @CommandLine.Option(
                names = {"-i", "--image"},
                paramLabel = "IMAGE",
                description = "the docker image. Default value maven project artifactId."
        )
        String image;

        /**
         * The docker repository
         */
        @CommandLine.Option(
                names = {"-r", "--repository"},
                paramLabel = "REPOSITORY",
                defaultValue = "docker.io",
                required = true,
                description = "the docker repository"
        )
        String repository;

        /**
         * Sets the maven project to release version.
         *
         * @return the exit code.
         * @throws Exception if the method fails.
         */
        @Override
        public Integer call() throws Exception {
            MavenProject project = getMavenProject();

            // set the docker image name.
            if (image == null || image.isEmpty()) {
                image = project.id.artifactId.value;
            }

            String imageName = repository + "/" + image;

            // execute the docker commands
            cmd("docker push " + imageName, "Error push docker image");

            logInfo("docker push " + imageName);
            return CommandLine.ExitCode.OK;
        }
    }
}
