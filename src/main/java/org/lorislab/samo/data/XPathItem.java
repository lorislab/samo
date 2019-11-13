package org.lorislab.samo.data;

import javax.xml.stream.XMLEventReader;
import javax.xml.stream.XMLInputFactory;
import javax.xml.stream.events.Characters;
import javax.xml.stream.events.StartElement;
import javax.xml.stream.events.XMLEvent;
import java.io.File;
import java.io.FileReader;
import java.util.HashMap;
import java.util.Map;
import java.util.Set;

public class XPathItem {

    public String value;

    public int begin;

    public int end;

    public XPathItem(int begin, int end, String value) {
        this.begin = begin;
        this.end = end;
        this.value = value;
    }

    public static Map<String, XPathItem> find(File file, Set<String> items) throws Exception {
        XMLInputFactory factory = XMLInputFactory.newFactory();
        XMLEventReader eventReader = factory.createXMLEventReader(new FileReader(file));
        Map<String, XPathItem> result = new HashMap<>();

        String xpath = "";
        while (eventReader.hasNext() && !items.isEmpty()) {
            XMLEvent event = eventReader.nextEvent();
            if (event.isStartElement()) {
                StartElement startElement = event.asStartElement();
                xpath = xpath + "/" + startElement.getName().getLocalPart();
            } else if (event.isEndElement()) {
                int index = xpath.lastIndexOf("/");
                if (index > -1) {
                    xpath = xpath.substring(0, index);
                }
            } else if (event.isCharacters()) {
                if (items.remove(xpath)) {
                    Characters chars = event.asCharacters();
                    String tmp = chars.getData();
                    int end = event.getLocation().getCharacterOffset() - 2;
                    int begin = end - tmp.length();
                    result.put(xpath, new XPathItem(begin, end, chars.getData()));
                }
            }
        }
        return result;
    }
}
