-- name: GlobalLeaderboard :many
-- Global reyting — faqat akkauntlar (mehmonlar emas), umumiy ball bo'yicha top 20.
SELECT u.username, COALESCE(SUM(gr.score),0)::double precision AS total_score,
       COUNT(*)::int AS games, COALESCE(SUM(gr.correct_cnt),0)::int AS correct
FROM game_results gr JOIN users u ON u.id = gr.user_id
WHERE u.username IS NOT NULL
GROUP BY u.id, u.username
ORDER BY total_score DESC
LIMIT 20;
