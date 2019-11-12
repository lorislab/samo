package org.lorislab.samo;

import org.lorislab.samo.docker.DockerTest;
import picocli.CommandLine;

@CommandLine.Command(
        name = "docker"
)
public class DockerCommand extends CommonCommand implements Runnable {
    @Override
    public void run() {
        try {
            DockerTest.release();
        } catch (Exception ex) {
            ex.printStackTrace();
        }
    }
}
