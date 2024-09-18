DROP TABLE IF EXISTS `url_map`;
CREATE TABLE `url_map`  (
                            `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT,
                            `long_url` varchar(250) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '长链',
                            `short_url` varchar(10) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '短链',
                            `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                            PRIMARY KEY (`id`) USING BTREE,
                            UNIQUE INDEX `idx_long_url_short_url`(`long_url` ASC, `short_url` ASC) USING BTREE,
                            INDEX `idx_short_url_long_url`(`short_url` ASC, `long_url` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 11 CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Dynamic;