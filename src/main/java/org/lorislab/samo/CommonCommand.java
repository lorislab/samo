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

import java.io.BufferedReader;
import java.io.File;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.util.concurrent.Callable;
import java.util.concurrent.CompletableFuture;

/**
 * The common abstract command.
 */
abstract class CommonCommand implements Callable<Integer> {

    /**
     * The snapshot suffix
     */
    static final String SNAPSHOT = "SNAPSHOT";

    /**
     * Is windows flag.
     */
    private static final boolean IS_WINDOWS = System.getProperty("os.name").toLowerCase().startsWith("windows");

    /**
     * Verbose flag.
     */
    @CommandLine.Option(
            names = {"-v", "--verbose"},
            defaultValue = "false",
            required = true,
            description = "the verbose output"
    )
    boolean verbose;

    /**
     * The maven project file.
     */
    @CommandLine.Option(
            names = {"-f", "--file"},
            paramLabel = "POM",
            defaultValue = "pom.xml",
            required = true,
            description = "the maven project file"
    )
    private File pom;

    /**
     * The log info message
     *
     * @param message the message.
     */
    void logInfo(String message) {
        System.out.println("SAMO >> " + message);
    }

    /**
     * The log verbose message.
     *
     * @param message the message.
     */
    void logVerbose(String message) {
        if (verbose) {
            System.out.println("SAMO >> " + message);
        }
    }

    /**
     * Gets the maven project.
     *
     * @return the maven project.
     * @throws Exception if the method fails.
     */
    MavenProject getMavenProject() throws Exception {
        logVerbose("Open maven project file: " + pom);
        MavenProject project = MavenProject.loadFromFile(pom);
        logVerbose("Project: " + project.id);
        return project;
    }

    /**
     * CLI run result object.
     */
    static class Return {
        /**
         * The exit value.
         */
        int exitValue = 0;
        /**
         * The response.
         */
        String response = "";
        /**
         * The error response.
         */
        String error = "";
    }

    /**
     * Call command line
     *
     * @param command
     * @param errorMessage
     * @return
     */
    Return cmd(String command, String errorMessage) {
        final Return result = new Return();
        try {
            ProcessBuilder pb = new ProcessBuilder();
            if (IS_WINDOWS) {
                pb.command("cmd", "/c", command);
            } else {
                pb.command("bash", "-c", command);
            }
            logVerbose("" + pb.command());
            final Process p = pb.start();
            CompletableFuture<String> out = readOutStream(p.getInputStream());
            CompletableFuture<String> err = readOutStream(p.getErrorStream());
            p.waitFor();

            result.response = out.get();
            result.error = err.get();
            result.exitValue = p.exitValue();

        } catch (Exception e) {
            throw new RuntimeException("Error processing command [" + command + "]", e);
        }

        if (result.exitValue != 0) {
            logInfo("Output: " + result.response);
            logVerbose("Error: " + result.error);
            logVerbose("Exit code: " + result.exitValue);
            throw new RuntimeException(errorMessage);
        }
        return result;
    }

    /**
     * The output stream reader.
     *
     * @param is the input stream.
     * @return the completable future to read the output stream.
     */
    private static CompletableFuture<String> readOutStream(InputStream is) {
        return CompletableFuture.supplyAsync(() -> {
            try (
                    InputStreamReader isr = new InputStreamReader(is);
                    BufferedReader br = new BufferedReader(isr);
            ) {
                StringBuilder res = new StringBuilder();
                String inputLine;
                while ((inputLine = br.readLine()) != null) {
                    res.append(inputLine);
                }
                return res.toString();
            } catch (Throwable e) {
                throw new RuntimeException("problem with executing program", e);
            }
        });
    }

}
