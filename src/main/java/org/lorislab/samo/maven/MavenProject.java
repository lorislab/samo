package org.lorislab.samo.maven;

import com.github.zafarkhaja.semver.Version;
import org.lorislab.samo.xml.XPathItem;

import java.io.File;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.util.Arrays;
import java.util.HashSet;
import java.util.Map;
import java.util.Set;

public class MavenProject {

    static String MAVEN_ARTIFACT_ID = "/project/artifactId";
    static String MAVEN_GROUP_ID = "/project/groupId";
    static String MAVEN_VERSION = "/project/version";
    static Set<String> ITEMS = new HashSet<>(Arrays.asList(MAVEN_GROUP_ID, MAVEN_ARTIFACT_ID, MAVEN_VERSION));

    public File file;

    public MavenProjectId id;

    public MavenProjectId parent;

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

    public boolean hasParent() {
        return parent != null;
    }

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
