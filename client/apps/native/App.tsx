import { useEffect } from "react";
import { SafeAreaView, View, Text, StyleSheet } from "react-native";
import { StatusBar } from "expo-status-bar";
import { useGame } from "@core/store";
import { configureCore } from "@core/config";
import "@core/i18n";
import { AuthScreen } from "./src/screens/AuthScreen";
import { HomeScreen } from "./src/screens/HomeScreen";

// Backend manzili — RN'da relative URL yo'q, absolute kerak.
// Android emulyator: 10.0.2.2 = host mashina; iOS simulyator: localhost.
configureCore({ apiBase: "http://10.0.2.2:8080", wsBase: "ws://10.0.2.2:8080" });

// Store-driven ekran routing — web App.tsx bilan AYNAN bir xil mantiq (@quizarena/core),
// faqat UI React Native komponentlari. Lobby/Play/Result ekranlari shu pattern bilan qo'shiladi.
export default function App() {
  const token = useGame((s) => s.token);
  const status = useGame((s) => s.status);
  const connect = useGame((s) => s.connect);

  useEffect(() => {
    if (token) connect();
  }, [token, connect]);

  return (
    <SafeAreaView style={styles.root}>
      <StatusBar style="dark" />
      <View style={styles.header}>
        <Text style={styles.brand}>QuizArena</Text>
        {token && status !== "online" && status !== "offline" && (
          <Text style={styles.status}>{status === "connecting" ? "Ulanmoqda…" : "Qayta ulanmoqda…"}</Text>
        )}
      </View>
      {!token ? <AuthScreen /> : <HomeScreen />}
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: "#ffffff" },
  header: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: "#e2e8f0",
  },
  brand: { fontSize: 16, fontWeight: "700", color: "#4f46e5" },
  status: { fontSize: 12, color: "#b45309" },
});
