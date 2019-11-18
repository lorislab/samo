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

import java.util.concurrent.Callable;

/**
 * The git command.
 */
@CommandLine.Command(name = "git",
        description = "Git commands",
        subcommands = {
                GitCommand.Branch.class,
                GitCommand.Hash.class
        }
)
public class GitCommand implements Callable<Integer> {

    /**
     * The command specification.
     */
    @CommandLine.Spec
    CommandLine.Model.CommandSpec spec;

    /**
     * Show help of the maven commands.
     *
     * @return the exit code.
     */
    @Override
    public Integer call() {
        spec.commandLine().usage(System.out);
        return CommandLine.ExitCode.OK;
    }

    /**
     * The maven version command.
     */
    @CommandLine.Command(name = "branch", description = "Show current branch")
    public static class Branch extends CommonCommand {

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
    @CommandLine.Command(name = "hash", description = "Show the git hash")
    public static class Hash extends CommonCommand {

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

        /**
         * Sets the maven project to release version.
         *
         * @return the exit code.
         */
        @Override
        public Integer call() {
            output(gitHash(length));
            return CommandLine.ExitCode.OK;
        }
    }

}
