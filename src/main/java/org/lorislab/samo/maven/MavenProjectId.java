package org.lorislab.samo.maven;

import org.lorislab.samo.xml.XPathItem;

public class MavenProjectId {
    public XPathItem artifactId;
    public XPathItem groupId;
    public XPathItem version;

    @Override
    public String toString() {
        return groupId.value + ":" + artifactId.value + ":" + version.value;
    }
}
