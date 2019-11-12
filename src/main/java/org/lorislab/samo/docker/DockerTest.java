package org.lorislab.samo.docker;

import org.lorislab.samo.cli.CliUtil;

import java.nio.file.Path;
import java.util.Optional;

public class DockerTest {

    public static void release() throws Exception {
        CliUtil.callCli(getDockerPath() + " pull quay.io/lorislab/fluh:0.3.0-8e160b2", "ERROR", true);
        CliUtil.callCli(getDockerPath() + " tag quay.io/lorislab/fluh:0.3.0-8e160b2 test/lorislab/fluh:0.3.0-8e160b2", "ERROR", true);
        CliUtil.callCli(getDockerPath() + " image rm test/lorislab/fluh:0.3.0-8e160b2", "ERROR", true);
    }

    static Path getDockerPath() {
        String helmExecutable = CliUtil.IS_WINDOWS ? "docker.exe" : "docker";
        Optional<Path> path = CliUtil.findInPath(helmExecutable);
        return path.orElseThrow(() -> new RuntimeException("docker executable is not found."));
    }

}
