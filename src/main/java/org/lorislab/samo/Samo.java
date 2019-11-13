package org.lorislab.samo;

import picocli.CommandLine;
import picocli.CommandLine.Command;

@Command(name = "samo",
        mixinStandardHelpOptions = true,
        version = "1.0",
        subcommands = {
                VersionCommand.class,
                CommandLine.HelpCommand.class
        }
)
public class Samo extends CommonCommand implements Runnable {

    @CommandLine.Spec
    CommandLine.Model.CommandSpec spec;

    public static void main(String[] args) {
        int exitCode = new CommandLine(new Samo()).execute(args);
        System.exit(exitCode);
    }

    public void run() {
        spec.commandLine().usage(System.out);
    }

}
