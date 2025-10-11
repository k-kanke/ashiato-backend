-- ユーザーテーブル
CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    profile_image_url TEXT,
    bio VARCHAR(500),
    is_banned BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);


-- ユーザー設定テーブル
CREATE TABLE IF NOT EXISTS user_settings (
    user_id UUID PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
    comment_on_my_pin BOOLEAN DEFAULT TRUE,
    friend_new_pin BOOLEAN DEFAULT TRUE,
    friend_request_received BOOLEAN DEFAULT TRUE,
    friend_request_accepted BOOLEAN DEFAULT TRUE
);


-- ピンテーブル (PostGISのジオメトリ型を使用)
CREATE TABLE IF NOT EXISTS pins (
    pin_id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    -- ジオメトリ型で緯度経度を保存し、SP-GiSTインデックスを張る
    location GEOMETRY(Point, 4326) NOT NULL, 
    content_text TEXT NOT NULL,
    media_url TEXT,
    privacy_setting VARCHAR(10) NOT NULL,
    status VARCHAR(10) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- ジオ検索を高速化するためのSP-GiSTインデックス
CREATE INDEX IF NOT EXISTS pins_location_idx ON pins USING SPGIST(location);


-- Comments (コメント/スレッド) テーブル
CREATE TABLE IF NOT EXISTS comments (
    comment_id UUID PRIMARY KEY,
    pin_id UUID NOT NULL REFERENCES pins(pin_id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    content_text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- スレッド表示を高速化するため、pin_idと作成日時でインデックスを付与
CREATE INDEX IF NOT EXISTS idx_comments_pin_created ON comments (pin_id, created_at ASC);


-- Friends (フレンド関係) テーブル
CREATE TABLE IF NOT EXISTS friends (
    user_a_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    user_b_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    PRIMARY KEY (user_a_id, user_b_id),
    status VARCHAR(10) NOT NULL,
    action_user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- ユーザーIDで効率的に検索するためのインデックス
CREATE INDEX IF NOT EXISTS idx_friends_user_a ON friends(user_a_id);
CREATE INDEX IF NOT EXISTS idx_friends_user_b ON friends(user_b_id);


-- Notifications (通知) テーブル
CREATE TABLE IF NOT EXISTS notifications (
    notification_id UUID PRIMARY KEY,
    recipient_user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    actor_user_id UUID REFERENCES users(user_id) ON DELETE SET NULL, 
    type VARCHAR(30) NOT NULL,
    related_entity_id UUID,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- 受信者IDと作成日時の降順でソート検索を高速化する複合インデックス
CREATE INDEX IF NOT EXISTS idx_notifications_recipient_created ON notifications (recipient_user_id, created_at DESC);