package org.lorislab.samo;

import picocli.CommandLine;

import java.io.File;

class MavenOptions {

    /**
     * The maven project file.
     */
    @CommandLine.Option(
            names = {"-p", "--pom"},
            paramLabel = "POM",
            defaultValue = "pom.xml",
            required = true,
            description = "the maven project file"
    )
    File pom;

}
