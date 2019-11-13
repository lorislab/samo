package org.lorislab.samo.data;

import org.lorislab.samo.data.XPathItem;

public class MavenProjectId {
    public XPathItem artifactId;
    public XPathItem groupId;
    public XPathItem version;

    @Override
    public String toString() {
        return groupId.value + ":" + artifactId.value + ":" + version.value;
    }
}
