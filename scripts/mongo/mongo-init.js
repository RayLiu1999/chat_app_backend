const targetEnv = (process.env.INIT_ENV || process.env.ENV || "development").toLowerCase();
const envSuffix = targetEnv.toUpperCase();

function getEnv(key, defaultValue = undefined) {
  const suffixStyle = process.env[`${key}_${envSuffix}`];
  const prefixStyle = process.env[`${envSuffix}_${key}`];
  const plain = process.env[key];

  if (suffixStyle !== undefined && suffixStyle !== "") return suffixStyle;
  if (prefixStyle !== undefined && prefixStyle !== "") return prefixStyle;
  if (plain !== undefined && plain !== "") return plain;
  return defaultValue;
}

function requireEnv(key, fallbackKey) {
  const value = getEnv(key) || (fallbackKey ? getEnv(fallbackKey) : undefined);
  if (!value) {
    throw new Error(`Missing required env: ${key}${fallbackKey ? ` (or ${fallbackKey})` : ""}`);
  }
  return value;
}

const dbName = getEnv("MONGO_DB_NAME", getEnv("MONGO_INITDB_DATABASE", "chat_app"));
const userName = requireEnv("MONGO_USERNAME", "MONGO_INITDB_APP_USERNAME");
const password = requireEnv("MONGO_PASSWORD", "MONGO_INITDB_APP_PASSWORD");

print(`[mongo-init] targetEnv=${targetEnv}, dbName=${dbName}`);

const appDb = db.getSiblingDB(dbName);

try {
  const existingUser = appDb.getUser(userName);
  if (!existingUser) {
    appDb.createUser({
      user: userName,
      pwd: password,
      roles: [{ role: "readWrite", db: dbName }],
    });
    print(`[mongo-init] user created: ${userName}`);
  } else {
    print(`[mongo-init] user exists, skip create: ${userName}`);
  }
} catch (error) {
  const message = String(error);
  if (message.includes("requires authentication") || message.includes("not authorized")) {
    print("[mongo-init] skip user management: insufficient privilege or unauthenticated");
  } else {
    throw error;
  }
}

const collections = [
  "users",
  "messages",
  "dm_rooms",
  "friends",
  "servers",
  "server_members",
  "channels",
  "files",
  "refresh_tokens",
];

const existingCollections = appDb.getCollectionNames();
collections.forEach((name) => {
  if (!existingCollections.includes(name)) {
    appDb.createCollection(name);
    print(`[mongo-init] collection created: ${name}`);
  }
});

const indexDefinitions = {
  users: [
    [{ username: 1 }, { unique: true }],
    [{ email: 1 }, { unique: true }],
    [{ is_online: 1 }, {}],
    [{ last_active_at: 1 }, {}],
  ],
  messages: [
    [{ room_id: 1, created_at: -1 }, {}],
    [{ sender_id: 1 }, {}],
    [{ room_type: 1 }, {}],
  ],
  dm_rooms: [
    [{ user_id: 1, chat_with_user_id: 1 }, { unique: true }],
    [{ room_id: 1 }, {}],
    [{ user_id: 1, is_hidden: 1 }, {}],
  ],
  friends: [
    [{ user_id: 1, friend_id: 1 }, { unique: true }],
    [{ user_id: 1, status: 1 }, {}],
  ],
  servers: [
    [{ name: 1 }, {}],
    [{ owner_id: 1 }, {}],
  ],
  server_members: [
    [{ server_id: 1, user_id: 1 }, { unique: true }],
    [{ user_id: 1 }, {}],
  ],
  channels: [
    [{ server_id: 1 }, {}],
    [{ server_id: 1, name: 1 }, { unique: true }],
  ],
  files: [
    [{ user_id: 1 }, {}],
    [{ file_type: 1 }, {}],
    [{ created_at: 1 }, {}],
  ],
  refresh_tokens: [
    [{ token: 1 }, { unique: true }],
    [{ user_id: 1 }, {}],
    [{ expires_at: 1 }, { expireAfterSeconds: 0 }],
  ],
};

Object.keys(indexDefinitions).forEach((collectionName) => {
  const collection = appDb.getCollection(collectionName);
  indexDefinitions[collectionName].forEach(([keys, options]) => {
    collection.createIndex(keys, options);
  });
});

print("[mongo-init] Database initialization completed successfully.");
