import { useState } from "react";
import { View, Text, TextInput, TouchableOpacity, StyleSheet } from "react-native";
import { useGame } from "@core/store";
import { api } from "@core/api";

// Mehmon yoki akkaunt bilan kirish — @core/api va @core/store web bilan AYNAN bir xil.
export function AuthScreen() {
  const setAuth = useGame((s) => s.setAuth);
  const setDisplayName = useGame((s) => s.setDisplayName);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [err, setErr] = useState("");

  async function guest() {
    try {
      setDisplayName(name || "O'yinchi");
      const r = await api.guest();
      setAuth(r.token, r.user);
    } catch (e) {
      setErr(e instanceof Error ? e.message : "xato");
    }
  }
  async function login() {
    try {
      const r = await api.login(email, password);
      setAuth(r.token, r.user);
    } catch (e) {
      setErr(e instanceof Error ? e.message : "xato");
    }
  }

  return (
    <View style={styles.box}>
      <Text style={styles.title}>Real-time bilim musobaqasi</Text>

      <TextInput style={styles.input} placeholder="Ismingiz (o'yinda)" value={name} onChangeText={setName} />
      <TouchableOpacity style={styles.primary} onPress={guest}>
        <Text style={styles.primaryText}>Mehmon sifatida o'ynash</Text>
      </TouchableOpacity>

      <Text style={styles.or}>— yoki —</Text>

      <TextInput
        style={styles.input}
        placeholder="Email"
        autoCapitalize="none"
        keyboardType="email-address"
        value={email}
        onChangeText={setEmail}
      />
      <TextInput style={styles.input} placeholder="Parol" secureTextEntry value={password} onChangeText={setPassword} />
      <TouchableOpacity style={styles.outline} onPress={login}>
        <Text style={styles.outlineText}>Kirish</Text>
      </TouchableOpacity>

      {err ? <Text style={styles.err}>{err}</Text> : null}
    </View>
  );
}

const styles = StyleSheet.create({
  box: { padding: 20, gap: 12 },
  title: { textAlign: "center", color: "#64748b", marginBottom: 8 },
  input: { borderWidth: 1, borderColor: "#cbd5e1", borderRadius: 10, paddingHorizontal: 12, paddingVertical: 10 },
  primary: { backgroundColor: "#4f46e5", borderRadius: 10, paddingVertical: 12, alignItems: "center" },
  primaryText: { color: "#ffffff", fontWeight: "600" },
  outline: { borderWidth: 1, borderColor: "#cbd5e1", borderRadius: 10, paddingVertical: 12, alignItems: "center" },
  outlineText: { color: "#334155", fontWeight: "600" },
  or: { textAlign: "center", color: "#94a3b8" },
  err: { color: "#dc2626", textAlign: "center" },
});
