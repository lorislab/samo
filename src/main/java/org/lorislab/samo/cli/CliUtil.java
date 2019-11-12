package org.lorislab.samo.cli;

import java.io.*;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.Optional;
import java.util.concurrent.CompletableFuture;
import java.util.regex.Pattern;
import java.util.stream.Stream;

public class CliUtil {

    public static boolean IS_WINDOWS = System.getProperty("os.name").toLowerCase().startsWith("windows");

    public static class Return {
        public int exitValue = 0;
        public String response = "";
        public String error = "";
    }

    public static Return callCli(String command, String errorMessage, final boolean verbose) {

        final Return result = new Return();
        try {
            if (verbose) {
                System.out.println("SAMO >> " + command);
            }
            ProcessBuilder pb = new ProcessBuilder();
            pb.command(command.split(" "));
            pb.start();
            final Process p = Runtime.getRuntime().exec(command);
            CompletableFuture<String> soutFut = readOutStream(p.getInputStream());
            CompletableFuture<String> serrFut = readOutStream(p.getErrorStream());
//            CompletableFuture<String> resultFut = soutFut.thenCombine(serrFut, (stdout, stderr) -> {
//                // print to current stderr the stderr of process and return the stdout
//                System.err.println(stderr);
//                return stdout;
//            });

//            new Thread(() -> {
//                BufferedReader input = new BufferedReader(new InputStreamReader(p.getInputStream()));
//                BufferedReader error = new BufferedReader(new InputStreamReader(p.getErrorStream()));
//                String inputLine;
//                String errorLine;
//                try {
//                    while ((inputLine = input.readLine()) != null) {
//                        if (verbose) {
//                            System.out.println(inputLine);
//                        }
//                        result.response = result.response + inputLine;
//                    }
//                    while ((errorLine = error.readLine()) != null) {
//                        System.err.println(errorLine);
//                        result.error = result.error + inputLine;
//                    }
//                } catch (IOException e) {
//                    throw new RuntimeException("Error reading the response for the command [" + command + "]", e);
//                }
//            }).start();
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
}
