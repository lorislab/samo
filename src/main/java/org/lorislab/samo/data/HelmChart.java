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
package org.lorislab.samo.data;

import java.io.File;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;

/**
 * The helm chart
 */
public class HelmChart {

    /**
     * The chart version attribute.
     */
    static String CHART_VERSION = "version";

    /**
     * The chart name attribute.
     */
    static String CHART_NAME = "name";

    /**
     * The chart version.
     */
    public YamlNode version;

    /**
     * The chart name.
     */
    public YamlNode name;

    /**
     * The helm chart file.
     */
    public File file;

    /**
     * Loads the helm chart from the file.
     *
     * @param file the helm chart file.
     * @return the helm chart.
     * @throws Exception if the method fails.
     */
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

    /**
     * Sets the helm chart new version.
     *
     * @param newVersion the new version to be set.
     */
    public void setVersion(String newVersion) {
        try {
            String data = new String(Files.readAllBytes(file.toPath()), StandardCharsets.UTF_8);
            data = data.substring(0, version.begin) + newVersion + data.substring(version.end);
            Files.write(file.toPath(), data.getBytes(StandardCharsets.UTF_8));
        } catch (Exception ex) {
            throw new RuntimeException("Error set new version " + version, ex);
        }
    }

    /**
     * {@inheritDoc}
     */
    @Override
    public String toString() {
        return name.value + ":" + version.value;
    }
}
