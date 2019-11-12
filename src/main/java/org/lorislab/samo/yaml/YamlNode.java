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
package org.lorislab.samo.yaml;

import org.yaml.snakeyaml.Yaml;
import org.yaml.snakeyaml.nodes.*;

import java.io.File;
import java.io.StringReader;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.util.*;

/**
 * The YAML node.
 */
public class YamlNode implements Iterable<YamlNode> {

    /**
     * The string value.
     */
    public String value;

    /**
     * The begin position in the file.
     */
    public int begin;

    /**
     * The end position in the file.
     */
    public int end;

    /**
     * The default constructor.
     */
    private YamlNode() {
        // default constructor
    }

    /**
     * The default constructor.
     *
     * @param value the value.
     * @param begin the begin position in the file.
     * @param end   the end position in the file.
     */
    private YamlNode(String value, int begin, int end) {
        this.value = value;
        this.begin = begin;
        this.end = end;
    }

    /**
     * The yaml node by the name.
     *
     * @param name the node name.
     * @return the yaml node.
     */
    public YamlNode get(String name) {
        return null;
    }

    /**
     * Gets the yaml node by the index.
     *
     * @param index the node index.
     * @return the yaml node.
     */
    public YamlNode get(int index) {
        return null;
    }

    /**
     * The size of the children
     *
     * @return the size of the children
     */
    public int size() {
        return 0;
    }

    /**
     * Returns {@code true} if the collection of children is empty.
     *
     * @return {@code true} if the collection of children is empty.
     */
    public boolean isEmpty() {
        return size() > 0;
    }

    /**
     * {@inheritDoc }
     */
    @Override
    public Iterator<YamlNode> iterator() {
        return Collections.emptyIterator();
    }

    /**
     * Gets the set of collection keys.
     *
     * @return the set of collection keys.
     */
    public Set<String> keys() {
        return Collections.emptySet();
    }

    /**
     * The list yaml node.
     */
    public static class ListYamlNode extends YamlNode {

        /**
         * The data.
         */
        private List<YamlNode> data;

        /**
         * The default constructor.
         *
         * @param data the data.
         */
        ListYamlNode(List<YamlNode> data) {
            this.data = data;
        }

        /**
         * {@inheritDoc }
         */
        @Override
        public YamlNode get(int index) {
            return data.get(index);
        }

        /**
         * {@inheritDoc }
         */
        @Override
        public int size() {
            return data.size();
        }

        /**
         * {@inheritDoc }
         */
        @Override
        public Iterator<YamlNode> iterator() {
            return data.iterator();
        }
    }

    /**
     * The map yaml node.
     */
    public static class MapYamlNode extends YamlNode {

        /**
         * The data.
         */
        private Map<String, YamlNode> data;

        /**
         * The default constructor.
         *
         * @param data the data.
         */
        MapYamlNode(Map<String, YamlNode> data) {
            this.data = data;
        }

        /**
         * {@inheritDoc }
         */
        @Override
        public YamlNode get(String name) {
            return data.get(name);
        }

        /**
         * {@inheritDoc }
         */
        @Override
        public int size() {
            return data.size();
        }

        /**
         * {@inheritDoc }
         */
        @Override
        public Set<String> keys() {
            return data.keySet();
        }
    }

    public static YamlNode parse(File file) throws Exception {
        String data = new String(Files.readAllBytes(file.toPath()), StandardCharsets.UTF_8);
        return YamlNode.parse(data);
    }
    /**
     * Parse the yaml data.
     *
     * @param data the data.
     * @return the yaml node.
     */
    public static YamlNode parse(String data) {
        Yaml yaml = new Yaml();
        Node node = yaml.compose(new StringReader(data));
        return value(node);
    }

    /**
     * Load the value node.
     *
     * @param node the node.
     * @return the corresponding yaml node.
     */
    private static YamlNode value(Node node) {
        switch (node.getNodeId()) {
            case mapping:
                return map((MappingNode) node);
            case scalar:
                ScalarNode sn = (ScalarNode) node;
                return new YamlNode(sn.getValue(), sn.getStartMark().getIndex(), sn.getEndMark().getIndex());
            case sequence:
                return array((SequenceNode) node);
            case anchor:
                return new YamlNode();
        }
        return new YamlNode();
    }

    /**
     * Loads the array yaml node.
     *
     * @param node the node.
     * @return the corresponding array yaml node.
     */
    private static YamlNode.ListYamlNode array(SequenceNode node) {
        List<YamlNode> result = new ArrayList<>();
        List<Node> children = node.getValue();
        for (Node child : children) {
            YamlNode value = value(child);
            result.add(value);
        }
        return new YamlNode.ListYamlNode(result);
    }

    /**
     * Loads the map yaml node.
     *
     * @param node the node.
     * @return the corresponding map yaml node.
     */
    private static YamlNode.MapYamlNode map(MappingNode node) {
        Map<String, YamlNode> result = new HashMap<>();
        List<NodeTuple> children = node.getValue();
        for (NodeTuple child : children) {
            YamlNode value = value(child.getValueNode());
            ScalarNode key = (ScalarNode) child.getKeyNode();
            result.put(key.getValue(), value);
        }
        return new YamlNode.MapYamlNode(result);
    }
}
