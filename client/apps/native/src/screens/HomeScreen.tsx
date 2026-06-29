import { View, Text, TouchableOpacity, StyleSheet } from "react-native";
import { useGame } from "@core/store";

// Asosiy ekran (poydevor). 🏆 O'ynash / 📚 O'rganish / 📊 Baholash tablari va
// Lobby/Play/Result ekranlari web sahifalari asosida RN'da qo'shiladi; mantiq
// (@quizarena/core store/api) o'zgarmaydi — faqat ko'rinish.
export function HomeScreen() {
  const displayName = useGame((s) => s.displayName);
  const status = useGame((s) => s.status);
  const logout = useGame((s) => s.logout);

  return (
    <View style={styles.box}>
      <Text style={styles.hi}>Salom, {displayName || "O'yinchi"} 👋</Text>
      <Text style={styles.conn}>Ulanish: {status}</Text>

      <View style={styles.tabs}>
        {["🏆 O'ynash", "📚 O'rganish", "📊 Baholash"].map((tab) => (
          <View key={tab} style={styles.tab}>
            <Text style={styles.tabText}>{tab}</Text>
          </View>
        ))}
      </View>

      <Text style={styles.note}>
        Ekranlar to'plami (Lobby, Play, Result, Practice, Assess) web bilan bir xil store
        oqimida qo'shiladi. Bu skelet @quizarena/core qayta ishlatilishini ko'rsatadi.
      </Text>

      <TouchableOpacity style={styles.outline} onPress={logout}>
        <Text style={styles.outlineText}>Chiqish</Text>
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  box: { padding: 20, gap: 14 },
  hi: { fontSize: 18, fontWeight: "700" },
  conn: { color: "#64748b" },
  tabs: { flexDirection: "row", gap: 8 },
  tab: { flex: 1, backgroundColor: "#eef2ff", borderRadius: 12, paddingVertical: 16, alignItems: "center" },
  tabText: { color: "#4338ca", fontWeight: "600" },
  note: { color: "#94a3b8", fontSize: 13, lineHeight: 19 },
  outline: { borderWidth: 1, borderColor: "#cbd5e1", borderRadius: 10, paddingVertical: 12, alignItems: "center" },
  outlineText: { color: "#334155", fontWeight: "600" },
});
