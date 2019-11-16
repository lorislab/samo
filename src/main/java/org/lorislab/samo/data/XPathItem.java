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

/**
 * The x-path item.
 */
public class XPathItem {

    /**
     * The element value.
     */
    public String value;

    /**
     * The begin in the file
     */
    int begin;

    /**
     * The end in the file.
     */
    int end;

    /**
     * Creates the xpath item.
     *
     * @param begin the begin in the file.
     * @param end   the end in the file.
     * @param value the element value.
     */
    private XPathItem(int begin, int end, String value) {
        this.begin = begin;
        this.end = end;
        this.value = value;
    }

    /**
     * Loads the xpath item from the xml file.
     *
     * @param file  the xml file.
     * @param items the xpath items.
     * @return the map of xpath items.
     * @throws Exception if the method fails.
     */
    static Map<String, XPathItem> find(File file, Set<String> items) throws Exception {
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
            } else if (event.isCharacters() && items.remove(xpath)) {
                Characters chars = event.asCharacters();
                String tmp = chars.getData();
                int end = event.getLocation().getCharacterOffset() - 2;
                int begin = end - tmp.length();
                result.put(xpath, new XPathItem(begin, end, chars.getData()));
            }
        }
        return result;
    }
}
