package org.lorislab.samo;

import org.lorislab.samo.data.MavenProject;
import picocli.CommandLine;

class DockerOptions {

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

}
