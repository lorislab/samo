package org.lorislab.samo;

import picocli.CommandLine;

import java.io.File;

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

    @CommandLine.Option(
            names = { "-e", "--env-version" },
            defaultValue = "SAMO_VERSION",
            required = true,
            description = "the version environment variable"
    )
    String versionVariable;
}
