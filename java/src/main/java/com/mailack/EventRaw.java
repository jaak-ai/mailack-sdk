package com.mailack;

public record EventRaw(byte[] data, String rawSha256) {}
