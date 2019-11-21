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

import picocli.CommandLine;
import picocli.CommandLine.Command;

import java.io.InputStream;
import java.util.Properties;

/**
 * The main command.
 */
@Command(name = "samo",
        mixinStandardHelpOptions = true,
        versionProvider = Samo.VersionProvider.class,
        subcommands = {
                MavenCommand.class,
                CreateCommand.class,
                DockerCommand.class,
                GitCommand.class,
                CommandLine.HelpCommand.class
        }
)
public class Samo extends SamoCommand {

    /**
     * Main method.
     *
     * @param args the command line arguments.
     */
    public static void main(String[] args) {
        int exitCode = new CommandLine(new Samo()).execute(args);
        System.exit(exitCode);
    }

    /**
     * The version provider.
     */
    public static class VersionProvider implements CommandLine.IVersionProvider {

        /**
         * {@inheritDoc}
         */
        @Override
        public String[] getVersion() throws Exception {
            Properties prop = new Properties();
            try (InputStream in = Samo.class.getResourceAsStream("/META-INF/maven/org.lorislab.samo/samo/pom.properties")) {
                prop.load(in);
            }
            return new String[]{prop.getProperty("version", "dev")};
        }
    }
}
