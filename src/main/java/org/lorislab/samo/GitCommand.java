package org.lorislab.samo;

import org.lorislab.samo.cli.CliUtil;
import picocli.CommandLine;

@CommandLine.Command(
        name = "git", addMethodSubcommands = true
)
public class GitCommand extends CommonCommand implements Runnable {

    @Override
    public void run() {
        try {
            CliUtil.Return r = CliUtil.callCli("git rev-parse --short=" + length + " HEAD", "Error git sha", verbose);
            System.out.println("Git hash: " + r.response);
        } catch (Exception ex) {
            ex.printStackTrace();
        }
    }

}
