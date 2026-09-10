package com.mailack;

public record MessageRaw(byte[] data, String canonicalHash) {}
