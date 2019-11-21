package org.lorislab.samo;

import picocli.CommandLine;

import java.io.File;

class MavenOptions {

    /**
     * The maven project file.
     */
    @CommandLine.Option(
            names = {"-f", "--file"},
            defaultValue = "pom.xml",
            required = true,
            description = "the maven project file"
    )
    File pom;

}
