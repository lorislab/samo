package org.lorislab.samo;

import org.lorislab.samo.helm.HelmChart;
import picocli.CommandLine;
import picocli.CommandLine.Command;

import java.io.File;

@Command( name = "helm", addMethodSubcommands = true)
public class HelmCommand extends CommonCommand implements Runnable {

    @CommandLine.Option(
            names = { "-c", "--chart" },
            paramLabel = "HELM_CHART",
            defaultValue = "Chart.yaml",
            required = true,
            description = "the helm chart file"
    )
    File chartFile;

    public void run() {
        System.out.println("Helm");
    }

    @Command(name = "release")
    public void release() {
        try {
            HelmChart helmChart = HelmChart.loadFromFile(chartFile);
            System.out.println("Project: " + helmChart);
            helmChart.release();
        } catch (Exception ex) {
            ex.printStackTrace();
        }
    }
}
