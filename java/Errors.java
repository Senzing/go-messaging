// Code generated by jtd-codegen for Java + Jackson v0.2.1

package com.senzing.schema;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonValue;
import java.util.List;

/**
 * A list of errors.  Usually a stack of errors.
 */
public class Errors {
    @JsonValue
    private List<Error> value;

    public Errors() {
    }

    @JsonCreator
    public Errors(List<Error> value) {
        this.value = value;
    }

    public List<Error> getValue() {
        return value;
    }

    public void setValue(List<Error> value) {
        this.value = value;
    }
}
