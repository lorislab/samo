package org.lorislab.samo;

import picocli.CommandLine;

import java.io.BufferedReader;
import java.io.File;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.Optional;
import java.util.concurrent.CompletableFuture;
import java.util.regex.Pattern;
import java.util.stream.Stream;

public abstract class CommonCommand implements Runnable {

    @CommandLine.Option(
            names = { "-l", "--length" },
            paramLabel = "LENGTH",
            defaultValue = "7",
            required = true,
            description = "the git sha length"
    )
    int length;

    @CommandLine.Option(
            names = { "-v", "--verbose" },
            defaultValue = "false",
            required = true,
            description = "the verbose output"
    )
    boolean verbose;

    @CommandLine.Option(
            names = { "-f", "--file" },
            paramLabel = "POM",
            defaultValue = "pom.xml",
            required = true,
            description = "the maven project file"
    )
    File pom;

    void logInfo(String message) {
        System.out.println(message);
    }

    void logVerbose(String message) {
        if (verbose) {
            System.out.println(message);
        }
    }

    public static boolean IS_WINDOWS = System.getProperty("os.name").toLowerCase().startsWith("windows");

    public static class Return {
        public int exitValue = 0;
        public String response = "";
        public String error = "";
    }

    Return callCli(String command, String errorMessage, final boolean verbose) {

        final Return result = new Return();
        try {
            logVerbose("SAMO >> " + command);
            ProcessBuilder pb = new ProcessBuilder();
            pb.command(command.split(" "));
            pb.start();
            final Process p = Runtime.getRuntime().exec(command);
            CompletableFuture<String> soutFut = readOutStream(p.getInputStream());
            CompletableFuture<String> serrFut = readOutStream(p.getErrorStream());
            p.waitFor();

            result.response =  soutFut.get();
            result.error =  serrFut.get();
            result.exitValue = p.exitValue();

        } catch (Exception e) {
            throw new RuntimeException("Error processing command [" + command + "]", e);
        }

        if (result.exitValue != 0) {
            throw new RuntimeException(errorMessage);
        }
        return result;
    }


    public static Optional<Path> findInPath(final String executable) {

        final String[] paths = getPathsFromEnvironmentVariables();
        return Stream.of(paths)
                .map(Paths::get)
                .map(path -> path.resolve(executable))
                .filter(Files::exists)
                .map(Path::toAbsolutePath)
                .findFirst();
    }

    public  static String[] getPathsFromEnvironmentVariables() {
        return System.getenv("PATH").split(Pattern.quote(File.pathSeparator));
    }

    static CompletableFuture<String> readOutStream(InputStream is) {
        return CompletableFuture.supplyAsync(() -> {
            try (
                    InputStreamReader isr = new InputStreamReader(is);
                    BufferedReader br = new BufferedReader(isr);
            ){
                StringBuilder res = new StringBuilder();
                String inputLine;
                while ((inputLine = br.readLine()) != null) {
                    res.append(inputLine);
                    //.append(System.lineSeparator())
                }
                return res.toString();
            } catch (Throwable e) {
                throw new RuntimeException("problem with executing program", e);
            }
        });
    }

    static Path getDockerPath() {
        String helmExecutable = IS_WINDOWS ? "docker.exe" : "docker";
        Optional<Path> path = findInPath(helmExecutable);
        return path.orElseThrow(() -> new RuntimeException("docker executable is not found."));
    }

    static Path getGitPath() {
        String helmExecutable = IS_WINDOWS ? "git.exe" : "git";
        Optional<Path> path = findInPath(helmExecutable);
        return path.orElseThrow(() -> new RuntimeException("git executable is not found."));
    }
}
