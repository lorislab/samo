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

import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;

/**
 * The maven command.
 */
@CommandLine.Command(name = "docker",
        description = "Docker version commands",
        showDefaultValues = true,
        mixinStandardHelpOptions = true,
        subcommands = {
                DockerCommand.Config.class,
                DockerCommand.Release.class,
                DockerCommand.Build.class,
                DockerCommand.Push.class
        }
)
class DockerCommand extends SamoCommand {

    @CommandLine.Command(name = "config",
            showDefaultValues = true,
            mixinStandardHelpOptions = true,
            description = "Configure the docker file.")
    public static class Config extends SamoCommand {

        /**
         * The docker config
         */
        @CommandLine.Option(
                names = {"-c", "--config"},
                paramLabel = "CONFIG",
                defaultValue = "${env:SAMO_DOCKER_CONFIG}",
                required = true,
                description = "the docker config. Env: SAMO_DOCKER_CONFIG"
        )
        String config;

        /**
         * The docker config
         */
        @CommandLine.Option(
                names = {"-j", "--config-file"},
                paramLabel = "CONFIG-FILE",
                defaultValue = "${env:SAMO_DOCKER_CONFIG_FILE:-~/.docker/config.json}",
                required = true,
                description = "the docker config file. Env: SAMO_DOCKER_CONFIG_FILE"
        )
        String configFile;

        @Override
        public Integer call() throws Exception {

            Path file = Paths.get(configFile);

            // create all directories
            Path dir = file.getParent();
            if (dir != null && !Files.exists(dir)) {
                Files.createDirectories(dir);
                debug("The docker config directory was created: %s", dir);
            }

            // write config to file
            Files.write(file, config.getBytes(StandardCharsets.UTF_8));
            debug("The docker config file was created. File: %s", file);
            return CommandLine.ExitCode.OK;
        }
    }

    /**
     * Release docker image.
     */
    @CommandLine.Command(name = "release",
            mixinStandardHelpOptions = true,
            showDefaultValues = true,
            description = "Release docker image")
    public static class Release extends SamoCommand {

        /**
         * The git options.
         */
        @CommandLine.Mixin
        GitOptions git;

        /**
         * The docker options.
         */
        @CommandLine.Mixin
        DockerOptions docker;

        /**
         * The maven options.
         */
        @CommandLine.Mixin
        MavenOptions maven;

        /**
         * Release docker image
         *
         * @return the exit code.
         * @throws Exception if the method fails.
         */
        @Override
        public Integer call() throws Exception {
            MavenProject project = getMavenProject(maven.pom);

            String hash = gitHash(git);

            Version version = Version.valueOf(project.id.version.value);
            Version pullVersion = version.setPreReleaseVersion(hash);

            // set the docker image name.
            updateImage(docker, project);

            String imageRelease = imageName(docker, version.getNormalVersion());
            String imagePull = imageName(docker, pullVersion.toString());

            // execute the docker commands
            cmd("docker pull " + imagePull, "Error pull docker image");
            cmd("docker tag " + imagePull + " " + imageRelease, "Error tag docker image");
            cmd("docker push " + imageRelease, "Error docker push image");

            info("Docker push new release image: %s", imageRelease);
            return CommandLine.ExitCode.OK;
        }
    }

    /**
     * Build docker image.
     */
    @CommandLine.Command(name = "build",
            mixinStandardHelpOptions = true,
            showDefaultValues = true,
            description = "Build docker image")
    public static class Build extends SamoCommand {

        /**
         * The docker options.
         */
        @CommandLine.Mixin
        DockerOptions docker;

        /**
         * The docker file.
         */
        @CommandLine.Option(
                names = {"-d", "--dockerfile"},
                defaultValue = "${env:SAMO_DOCKER_DOCKERFILE:-src/main/docker/Dockerfile}",
                description = "the docker file. Env: SAMO_DOCKER_DOCKERFILE"
        )
        String dockerfile;

        /**
         * The docker context.
         */
        @CommandLine.Option(
                names = {"-c", "--context"},
                required = true,
                defaultValue = ".",
                description = "the docker build context"
        )
        String context;

        /**
         * Create docker image tag for the branch.
         */
        @CommandLine.Option(
                names = {"-b", "--no-branch"},
                defaultValue = "true",
                required = true,
                negatable = true,
                description = "tag the docker image with a branch name"
        )
        boolean branch;

        /**
         * Create docker image tag for the branch.
         */
        @CommandLine.Option(
                names = {"-l", "--no-latest"},
                defaultValue = "true",
                negatable = true,
                required = true,
                description = "tag the docker image with a branch name"
        )
        boolean latest;

        /**
         * The maven options.
         */
        @CommandLine.Mixin
        MavenOptions maven;

        /**
         * Build docker image.
         *
         * @return the exit code.
         * @throws Exception if the method fails.
         */
        @Override
        public Integer call() throws Exception {
            MavenProject project = getMavenProject(maven.pom);

            // set the docker image name.
            updateImage(docker, project);

            StringBuilder log = new StringBuilder();
            log.append("Docker build new images [");

            StringBuilder sb = new StringBuilder();
            sb.append("docker build");

            String imageName = imageName(docker, project.id.version.value);
            sb.append(" -t ").append(imageName);
            log.append(imageName);

            if (branch) {
                String branchName = gitBranch();
                String branchTag = imageName(docker, branchName);
                sb.append(" -t ").append(branchTag);
                log.append(",").append(branchTag);
            }
            if (latest) {
                String branchTag = imageName(docker, "latest");
                sb.append(" -t ").append(branchTag);
                log.append(",").append(branchTag);
            }
            if (dockerfile != null && !dockerfile.isEmpty()) {
                sb.append(" -f ").append(dockerfile);
            }
            sb.append(" ").append(context);
            log.append("]");

            // execute the docker commands
            cmd(sb.toString(), "Error build docker image");

            info(log.toString());
            return CommandLine.ExitCode.OK;
        }
    }

    /**
     * Docker push command
     */
    @CommandLine.Command(name = "push",
            mixinStandardHelpOptions = true,
            showDefaultValues = true,
            description = "Push docker image")
    public static class Push extends SamoCommand {

        /**
         * The docker options.
         */
        @CommandLine.Mixin
        DockerOptions docker;

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

            // set the docker image name.
            updateImage(docker, project);

            String imageName = imageName(docker, null);

            // execute the docker commands
            cmd("docker push " + imageName, "Error push docker image");

            info("docker push %s", imageName);
            return CommandLine.ExitCode.OK;
        }
    }

    /**
     * Gets the image name.
     *
     * @param options the docker options.
     * @param tag     the image tag.
     * @return the corresponding image name.
     */
    static String imageName(DockerOptions options, String tag) {
        String tmp = options.repository + "/" + options.image;
        if (tag != null) {
            tmp = tmp + ":" + tag;
        }
        return tmp;
    }

    /**
     * Update the docker image to maven project artifact ID if there is no options image.
     *
     * @param options the docker options.
     * @param project the maven project.
     */
    static void updateImage(DockerOptions options, MavenProject project) {
        if (options.image == null || options.image.isEmpty()) {
            options.image = project.id.artifactId.value;
        }
    }

    static class DockerOptions {

        /**
         * The docker repository
         */
        @CommandLine.Option(
                names = {"-i", "--image"},
                description = "the docker image. Default value maven project artifactId."
        )
        String image;

        /**
         * The docker repository
         */
        @CommandLine.Option(
                names = {"-r", "--repository"},
                defaultValue = "${env:SAMO_DOCKER_REPOSITORY:-docker.io}",
                required = true,
                description = "the docker repository. Env: SAMO_DOCKER_REPOSITORY"
        )
        String repository;

    }
}
