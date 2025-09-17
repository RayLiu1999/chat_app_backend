// 使用 process.env 來取得環境變數
const dbName = process.env.MONGO_DB_NAME;
const userName = process.env.MONGO_USERNAME;
const password = process.env.MONGO_PASSWORD;

// 進入指定的資料庫
db = db.getSiblingDB(dbName);

// 創建go應用用戶
db.createUser({
  user: userName,
  pwd: password,
  roles: [
    {
      role: 'readWrite',
      db: dbName
    }
  ]
});

// 建立collections
db.createCollection("users");
db.createCollection("messages");
db.createCollection("dm_rooms");
db.createCollection("friends");
db.createCollection("servers");
db.createCollection("server_members");
db.createCollection("channels");
db.createCollection("files");
db.createCollection("refresh_tokens");

// 創建集合和索引
print('Creating collections and indexes...');

// 用戶集合索引
// db.users.createIndex({ "username": 1 }, { unique: true });
// db.users.createIndex({ "email": 1 }, { unique: true });
// db.users.createIndex({ "is_online": 1 });
// db.users.createIndex({ "last_active_at": 1 });

// // 訊息集合索引
// db.messages.createIndex({ "room_id": 1, "created_at": -1 });
// db.messages.createIndex({ "sender_id": 1 });
// db.messages.createIndex({ "room_type": 1 });

// // DM 房間集合索引
// db.dm_rooms.createIndex({ "user_id": 1, "chat_with_user_id": 1 }, { unique: true });
// db.dm_rooms.createIndex({ "room_id": 1 });
// db.dm_rooms.createIndex({ "user_id": 1, "is_hidden": 1 });

// // 好友集合索引
// db.friends.createIndex({ "user_id": 1, "friend_id": 1 }, { unique: true });
// db.friends.createIndex({ "user_id": 1, "status": 1 });

// // 伺服器集合索引
// db.servers.createIndex({ "name": 1 });
// db.servers.createIndex({ "owner_id": 1 });

// // 伺服器成員集合索引
// db.server_members.createIndex({ "server_id": 1, "user_id": 1 }, { unique: true });
// db.server_members.createIndex({ "user_id": 1 });

// // 頻道集合索引
// db.channels.createIndex({ "server_id": 1 });
// db.channels.createIndex({ "server_id": 1, "name": 1 }, { unique: true });

// // 檔案集合索引
// db.files.createIndex({ "user_id": 1 });
// db.files.createIndex({ "file_type": 1 });
// db.files.createIndex({ "created_at": 1 });

// // Refresh Token 集合索引
// db.refresh_tokens.createIndex({ "token": 1 }, { unique: true });
// db.refresh_tokens.createIndex({ "user_id": 1 });
// db.refresh_tokens.createIndex({ "expires_at": 1 }, { expireAfterSeconds: 0 });

print('Database initialization completed successfully!');

// 插入測試數據（可選）
// if (db.users.countDocuments() === 0) {
//   print('Inserting test data...');
  
//   // 創建測試用戶
//   db.users.insertMany([
//     {
//       _id: ObjectId(),
//       username: "testuser1",
//       email: "test1@example.com",
//       password: "$2a$10$example_hashed_password_1",
//       nickname: "測試用戶1",
//       is_online: false,
//       created_at: new Date(),
//       updated_at: new Date()
//     },
//     {
//       _id: ObjectId(),
//       username: "testuser2", 
//       email: "test2@example.com",
//       password: "$2a$10$example_hashed_password_2",
//       nickname: "測試用戶2",
//       is_online: false,
//       created_at: new Date(),
//       updated_at: new Date()
//     }
//   ]);
  
//   print('Test data inserted successfully!');
// }
