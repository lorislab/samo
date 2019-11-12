package org.lorislab.samo.helm;

import com.github.zafarkhaja.semver.Version;
import org.lorislab.samo.yaml.YamlNode;

import java.io.File;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;

public class HelmChart {

    static String CHART_VERSION = "version";

    static String CHART_NAME = "name";

    public YamlNode version;

    public YamlNode name;

    public File file;

    public static HelmChart loadFromFile(File file) throws Exception {
        if (file == null || !file.exists() || file.isDirectory()) {
            return null;
        }

        YamlNode root = YamlNode.parse(file);

        HelmChart chart = new HelmChart();
        chart.file = file;
        chart.version = root.get(CHART_VERSION);
        chart.name = root.get(CHART_NAME);
        return chart;
    }

    public void release() {
        Version tmp = Version.valueOf(version.value);
        String releaseVersion = tmp.getNormalVersion();
        setVersion(releaseVersion);
        System.out.println("Change version from " + version.value + " to " + releaseVersion);
    }

    public void setVersion(String newVersion) {
        try {
            String data = new String(Files.readAllBytes(file.toPath()), StandardCharsets.UTF_8);
            data = data.substring(0, version.begin) + newVersion + data.substring(version.end);
            Files.write(file.toPath(), data.getBytes(StandardCharsets.UTF_8));
        } catch (Exception ex) {
            throw new RuntimeException("Error set new version " + version, ex);
        }
    }

    @Override
    public String toString() {
        return name.value + ":" + version.value;
    }
}
