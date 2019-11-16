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
import java.util.Arrays;
import java.util.HashSet;
import java.util.Map;
import java.util.Set;

/**
 * The maven project.
 */
public class MavenProject {

    /**
     * The xpath artifact ID.
     */
    static String MAVEN_ARTIFACT_ID = "/project/artifactId";

    /**
     * The xpath group ID.
     */
    static String MAVEN_GROUP_ID = "/project/groupId";

    /**
     * The xpath version.
     */
    static String MAVEN_VERSION = "/project/version";

    /**
     * The list of default xpath items to load.
     */
    static Set<String> ITEMS = new HashSet<>(Arrays.asList(MAVEN_GROUP_ID, MAVEN_ARTIFACT_ID, MAVEN_VERSION));

    /**
     * The maven project file.
     */
    public File file;

    /**
     * The maven project ID.
     */
    public MavenProjectId id;

    /**
     * The maven project parent project ID.
     */
    public MavenProjectId parent;

    /**
     * Loads the maven project from file.
     *
     * @param file the maven project file.
     * @return the maven project.
     * @throws Exception if the method fails.
     */
    public static MavenProject loadFromFile(File file) throws Exception {
        if (file == null || !file.exists() || file.isDirectory()) {
            return null;
        }

        Map<String, XPathItem> items = XPathItem.find(file, ITEMS);
        if (items.isEmpty()) {
            return null;
        }
        MavenProject project = new MavenProject();
        project.file = file;
        project.id = new MavenProjectId();
        project.id.groupId = items.get(MAVEN_GROUP_ID);
        project.id.artifactId = items.get(MAVEN_ARTIFACT_ID);
        project.id.version = items.get(MAVEN_VERSION);
        return project;
    }

    /**
     * Sets the maven project new version.
     *
     * @param version the version to be set.
     */
    public void setVersion(String version) {
        try {
            String data = new String(Files.readAllBytes(file.toPath()), StandardCharsets.UTF_8);
            data = data.substring(0, id.version.begin) + version + data.substring(id.version.end);
            Files.write(file.toPath(), data.getBytes(StandardCharsets.UTF_8));
        } catch (Exception ex) {
            throw new RuntimeException("Error set version " + version, ex);
        }
    }

}
