package org.lorislab.samo;

import com.github.zafarkhaja.semver.Version;
import org.lorislab.samo.data.MavenProject;
import picocli.CommandLine;

import java.io.BufferedReader;
import java.io.File;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.util.concurrent.Callable;
import java.util.concurrent.CompletableFuture;

class SamoCommand implements Callable<Integer> {

    /**
     * Is windows flag.
     */
    private static final boolean IS_WINDOWS = System.getProperty("os.name").toLowerCase().startsWith("windows");

    /**
     * The snapshot suffix
     */
    static final String SNAPSHOT = "SNAPSHOT";

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
     * Show help of the tool.
     *
     * @return the exit code.
     */
    public Integer call() throws Exception {
        CommandLine.usage(this, System.out);
        return CommandLine.ExitCode.OK;
    }


    /**
     * The log info message
     *
     * @param message the message.
     */
    void output(String message, Object... params) {
        System.out.printf(message, params);
    }

    /**
     * The log verbose message.
     *
     * @param message the message.
     */
    void info(String message, Object... params) {
        output("SAMO >> " + message + "\n", params);
    }

    /**
     * The log verbose message.
     *
     * @param message the message.
     */
    void debug(String message, Object... params) {
        if (verbose) {
            info(message, params);
        }
    }

    void setMavenVersion(MavenProject project, String version) {
        project.setVersion(version);
        info("Change version from " + project.id.version.value + " to " + version + " in the file: " + project.file);
    }

    /**
     * Gets the maven project.
     *
     * @return the maven project.
     * @throws Exception if the method fails.
     */
    MavenProject getMavenProject(File pom) throws Exception {
        debug("Open maven project file: %s", pom);
        MavenProject project = MavenProject.loadFromFile(pom);
        debug("Maven project ID: %s", project.id);
        return project;
    }

    /**
     * Gets the pre-release version of the maven project.
     *
     * @param project    the maven project.
     * @param preRelease the pre release tag.
     * @return the corresponding pre release version.
     */
    String preReleaseVersion(MavenProject project, String preRelease) {
        Version version = Version.valueOf(project.id.version.value);
        version = version.setPreReleaseVersion(preRelease);
        return version.toString();
    }


    /**
     * Gets the git hash commit.
     *
     * @param options the git options.
     * @return the corresponding git hash.
     */
    String gitHash(GitOptions options) {
        Return r = cmd("git rev-parse --short=" + options.length + " HEAD", "Error git hash", false);
        debug("Git hash: %s", r.response);
        return r.response;
    }

    /**
     * Gets the branch name.
     *
     * @return the git branch.
     */
    String gitBranch() {
        if (isGitHub()) {
            String b = System.getenv("GITHUB_REF");
            if (b != null && !b.isEmpty()) {
                debug("Github branch: %s", b);
                return b.replace("refs/heads/", "");
            }
        }
        if (isGitLab()) {
            return System.getenv("CI_COMMIT_REF_NAME");
        }
        Return r = cmd("git rev-parse --abbrev-ref HEAD", "Error git branch name", false);
        debug("Git branch: %s", r.response);
        return r.response;
    }

    /**
     * Returns {@code true} if the pipeline is github actions pipeline
     *
     * @return {@code true} if the pipeline is github actions pipeline
     */
    boolean isGitHub() {
        String tmp = System.getenv("GITHUB_ACTIONS");
        boolean result = Boolean.parseBoolean(tmp);
        debug("? Github: %s", result);
        return result;
    }

    /**
     * Returns {@code true} if the pipeline is gitlab-ci pipeline.
     *
     * @return {@code true} if the pipeline is gitlab-ci pipeline.
     */
    boolean isGitLab() {
        String tmp = System.getenv("GITLAB_CI");
        boolean result = Boolean.parseBoolean(tmp);
        debug("? GitLab: %s", result);
        return result;
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
     * @param command      the command.
     * @param errorMessage the error message.
     * @return the command line response.
     */
    Return cmd(String command, String errorMessage) {
        return cmd(command, errorMessage, true);
    }

    /**
     * Call command line
     *
     * @param command      the command.
     * @param errorMessage the error message.
     * @param newline      add the new line character to the output stream.
     * @return the command line response.
     */
    Return cmd(String command, String errorMessage, boolean newline) {
        final Return result = new Return();
        try {
            ProcessBuilder pb = new ProcessBuilder();
            if (IS_WINDOWS) {
                pb.command("cmd", "/c", command);
            } else {
                pb.command("bash", "-c", command);
            }
            debug("" + pb.command());
            final Process p = pb.start();
            CompletableFuture<String> out = readOutStream(p.getInputStream(), newline);
            CompletableFuture<String> err = readOutStream(p.getErrorStream(), newline);
            p.waitFor();

            result.response = out.get();
            result.error = err.get();
            result.exitValue = p.exitValue();

        } catch (Exception e) {
            throw new RuntimeException("Error processing command [" + command + "]", e);
        }

        if (result.exitValue != 0) {
            debug("Output: %s", result.response);
            debug("Error: %s", result.error);
            debug("Exit code: %s", result.exitValue);
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
    private static CompletableFuture<String> readOutStream(InputStream is, boolean newline) {
        return CompletableFuture.supplyAsync(() -> {
            try (
                    InputStreamReader isr = new InputStreamReader(is);
                    BufferedReader br = new BufferedReader(isr);
            ) {
                StringBuilder res = new StringBuilder();
                String inputLine;
                while ((inputLine = br.readLine()) != null) {
                    res.append(inputLine);
                    if (newline) {
                        res.append("\n");
                    }
                }
                return res.toString();
            } catch (Throwable e) {
                throw new RuntimeException("problem with executing program", e);
            }
        });
    }

}
