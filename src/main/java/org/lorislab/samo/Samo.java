package org.lorislab.samo;

import picocli.CommandLine;
import picocli.CommandLine.Command;

import java.util.concurrent.Callable;

@Command(name = "samo",
        mixinStandardHelpOptions = true,
        version = "1.0",
        subcommands = {
                VersionCommand.class,
                CommandLine.HelpCommand.class
        }
)
public class Samo extends CommonCommand implements Callable<Integer> {

    @CommandLine.Spec
    CommandLine.Model.CommandSpec spec;

    public static void main(String[] args) {
        int exitCode = new CommandLine(new Samo()).execute(args);
        System.exit(exitCode);
    }

    public Integer call() {
        spec.commandLine().usage(System.out);
        return CommandLine.ExitCode.OK;
    }

}
