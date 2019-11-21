package org.lorislab.samo;

import picocli.CommandLine;

class GitOptions {

    /**
     * The length of the git sha
     */
    @CommandLine.Option(
            names = {"-l", "--length"},
            paramLabel = "LENGTH",
            defaultValue = "7",
            required = true,
            description = "the git hash length"
    )
    int length;

}
