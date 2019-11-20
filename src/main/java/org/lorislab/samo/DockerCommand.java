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
         * The docker image
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
                defaultValue = "${env:SAMO_DOCKER_REPOSITORY:-docker.io}",
                required = true,
                description = "the docker repository. Env: SAMO_DOCKER_REPOSITORY"
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

            String hash = gitHash(length);

            Version version = Version.valueOf(project.id.version.value);
            Version pullVersion = version.setPreReleaseVersion(hash);

            // set the docker image name.
            if (image == null || image.isEmpty()) {
                image = project.id.artifactId.value;
            }
            String imageRelease = repository + "/" + image + ":" + version.getNormalVersion();
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
                defaultValue = "${env:SAMO_DOCKER_REPOSITORY:-docker.io}",
                required = true,
                description = "the docker repository. Env: SAMO_DOCKER_REPOSITORY"
        )
        String repository;

        /**
         * The docker file.
         */
        @CommandLine.Option(
                names = {"-d", "--dockerfile"},
                paramLabel = "DOCKERFILE",
                defaultValue = "${env:SAMO_DOCKER_DOCKERFILE}",
                description = "the docker file. Env: SAMO_DOCKER_DOCKERFILE"
        )
        String dockerfile;

        /**
         * The docker context.
         */
        @CommandLine.Option(
                names = {"-c", "--context"},
                paramLabel = "CONTEXT",
                required = true,
                defaultValue = ".",
                description = "the docker build context"
        )
        String context;

        /**
         * Verbose flag.
         */
        @CommandLine.Option(
                names = {"-b", "--branch"},
                defaultValue = "true",
                required = true,
                description = "tag the docker image with a branch name"
        )
        boolean branch;

        /**
         * The docker password.
         */
        @CommandLine.Option(
                names = {"-p", "--password"},
                paramLabel = "PASSWORD",
                defaultValue = "${env:SAMO_DOCKER_PASSWORD}",
                description = "the docker login password"
        )
        String password;

        /**
         * The docker password.
         */
        @CommandLine.Option(
                names = {"-u", "--username"},
                paramLabel = "USERNAME",
                defaultValue = "${env:SAMO_DOCKER_USERNAME}",
                description = "the docker login username"
        )
        String username;

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
                String branchName = gitBranch();
                branchTag = repository + "/" + image + ":" + branchName;
                sb.append(" -t ").append(branchTag);
                log.append(",").append(branchTag);
            }
            if (dockerfile != null && !dockerfile.isEmpty()) {
                sb.append(" -f ").append(dockerfile);
            }
            sb.append(" ").append(context);
            log.append("]");

            // execute the docker commands
            if (password != null && username != null) {
                cmd("docker login -p " + password + " -u " + username, "Error docker login");
            }
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
                defaultValue = "${env:SAMO_DOCKER_REPOSITORY:-docker.io}",
                required = true,
                description = "the docker repository. Env: SAMO_DOCKER_REPOSITORY"
        )
        String repository;

        /**
         * The docker password.
         */
        @CommandLine.Option(
                names = {"-p", "--password"},
                paramLabel = "PASSWORD",
                defaultValue = "${env:SAMO_DOCKER_PASSWORD}",
                description = "the docker login password"
        )
        String password;

        /**
         * The docker password.
         */
        @CommandLine.Option(
                names = {"-u", "--username"},
                paramLabel = "USERNAME",
                defaultValue = "${env:SAMO_DOCKER_USERNAME}",
                description = "the docker login username"
        )
        String username;

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
            if (password != null && username != null) {
                cmd("docker login -p " + password + " -u " + username, "Error docker login");
            }
            cmd("docker push " + imageName, "Error push docker image");

            logInfo("docker push " + imageName);
            return CommandLine.ExitCode.OK;
        }
    }
}
