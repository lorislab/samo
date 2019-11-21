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

/**
 * The git command.
 */
@CommandLine.Command(name = "git",
        mixinStandardHelpOptions = true,
        description = "Git commands",
        showDefaultValues = true,
        subcommands = {
                GitCommand.Branch.class,
                GitCommand.Hash.class
        }
)
class GitCommand extends SamoCommand {

    /**
     * The maven version command.
     */
    @CommandLine.Command(name = "branch",
            mixinStandardHelpOptions = true,
            description = "Show current branch")
    public static class Branch extends SamoCommand {

        /**
         * Returns the current version of the maven project.
         *
         * @return the exit code.
         */
        @Override
        public Integer call() {
            output(gitBranch());
            return CommandLine.ExitCode.OK;
        }
    }

    /**
     * Gets the git hash.
     */
    @CommandLine.Command(name = "hash",
            mixinStandardHelpOptions = true,
            showDefaultValues = true,
            description = "Show the git hash")
    public static class Hash extends SamoCommand {

        /**
         * The git options.
         */
        @CommandLine.Mixin
        GitOptions git;

        /**
         * Sets the maven project to release version.
         *
         * @return the exit code.
         */
        @Override
        public Integer call() {
            output(gitHash(git));
            return CommandLine.ExitCode.OK;
        }
    }

}
