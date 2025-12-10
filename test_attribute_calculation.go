package main

import (
	"fmt"
	"strconv"
	"strings"

	"ggo/models"
)

func main() {
	// 模拟用户ID
	userID := uint(9)

	// 定义属性汇总结构
	attributes := map[string]interface{}{
		"hp":              0,
		"attack":          0,
		"attack_speed":    1.0,
		"move_speed":      0,
		"bullet_speed":    0,
		"drain":           0,
		"critical":        0,
		"dodge":           0,
		"instant_kill":    0,
		"recovery":        0,
		"trajectory":      0,
		"critical_rate":   0.0,
		"critical_damage": 1.5,
		"atk_type":        0,
	}

	// 模拟装备数据（根据用户提供的示例）
	equippedItems := []models.UserEquipment{
		{
			ID:           10,
			UserID:       9,
			EquipmentID:  10,
			IsEquipped:   true,
			Position:     "equipped",
			EnhanceLevel: 0,
			EquipmentTemplate: models.EquipmentTemplate{
				ID:           10,
				Name:         "轩辕剑",
				Level:        6,
				Slot:         "weapon",
				HP:           0,
				Attack:       300,
				AttackSpeed:  10,
				MoveSpeed:    0,
				BulletSpeed:  0,
				Drain:        0,
				Critical:     0,
				Dodge:        0,
				InstantKill:  0,
				Recovery:     0,
				Trajectory:   0,
				ImageURL:     "https://czrimg.godqb.com/game/zb/wq/6.jpg",
				Description:  "1",
				IsActive:     true,
				CreatedAt:    0,
				UpdatedAt:    0,
			},
			AdditionalAttrs: []models.EquipmentAdditionalAttr{
				{
					ID:              9,
					UserEquipmentID: 10,
					AttrType:        "damage_reduction",
					AttrName:        "减伤",
					AttrValue:       "2.8%",
					CreatedAt:       1765274069,
					UpdatedAt:       1765274069,
				},
			},
		},
	}

	// 模拟皮肤数据（根据用户提供的示例）
	activeSkin := models.UserSkin{
		ID:       10,
		UserID:   9,
		SkinID:   1,
		IsActive: true,
		Skin: models.Skin{
			ID:             1,
			Name:           "修女",
			Attack:         5,
			HP:             50,
			AtkType:        0,
			AtkSpeed:       0,
			CriticalRate:   0,
			CriticalDamage: 0,
			// 其他皮肤属性省略
		},
	}

	fmt.Println("=== 模拟属性计算测试 ===")
	fmt.Printf("用户ID: %d\n\n", userID)

	// 计算装备属性总和
	fmt.Println("1. 计算装备属性:")
	for _, item := range equippedItems {
		fmt.Printf("   装备: %s\n", item.EquipmentTemplate.Name)
		fmt.Printf("   - 基础攻击: %d\n", item.EquipmentTemplate.Attack)
		fmt.Printf("   - 基础攻击速度: %f\n", item.EquipmentTemplate.AttackSpeed)
		fmt.Printf("   - 基础HP: %d\n", item.EquipmentTemplate.HP)

		// 基础属性
		attributes["hp"] = attributes["hp"].(int) + item.EquipmentTemplate.HP
		attributes["attack"] = attributes["attack"].(int) + item.EquipmentTemplate.Attack
		attributes["attack_speed"] = attributes["attack_speed"].(float64) + item.EquipmentTemplate.AttackSpeed
		attributes["move_speed"] = attributes["move_speed"].(int) + item.EquipmentTemplate.MoveSpeed
		attributes["bullet_speed"] = attributes["bullet_speed"].(int) + item.EquipmentTemplate.BulletSpeed
		attributes["drain"] = attributes["drain"].(int) + item.EquipmentTemplate.Drain
		attributes["critical"] = attributes["critical"].(int) + item.EquipmentTemplate.Critical
		attributes["dodge"] = attributes["dodge"].(int) + item.EquipmentTemplate.Dodge
		attributes["instant_kill"] = attributes["instant_kill"].(int) + item.EquipmentTemplate.InstantKill
		attributes["recovery"] = attributes["recovery"].(int) + item.EquipmentTemplate.Recovery
		attributes["trajectory"] = attributes["trajectory"].(int) + item.EquipmentTemplate.Trajectory

		// 附加属性处理
		for _, attr := range item.AdditionalAttrs {
			fmt.Printf("   - 附加属性: %s(%s): %s\n", attr.AttrName, attr.AttrType, attr.AttrValue)

			if attr.AttrType == "damage_reduction" {
				// 减伤属性特殊处理，暂时存储为字符串
				if _, exists := attributes["damage_reduction"]; exists {
					// 如果已经存在减伤属性，将两个字符串合并
					currentDR := attributes["damage_reduction"].(string)
					attributes["damage_reduction"] = currentDR + " + " + attr.AttrValue
				} else {
					attributes["damage_reduction"] = attr.AttrValue
				}
			} else if attr.AttrType == "enhance" {
				// 处理enhance类型的特殊稀有属性
				cleanValue := attr.AttrValue
				if strings.Contains(cleanValue, "%") {
					cleanValue = strings.ReplaceAll(cleanValue, "%", "")
				}
				if strings.Contains(cleanValue, "秒杀") {
					cleanValue = strings.ReplaceAll(cleanValue, "秒杀", "")
				}

				// 根据AttrName判断属性类型并累加
				switch attr.AttrName {
				case "暴食": // 增加攻速
					if val, err := strconv.ParseFloat(cleanValue, 64); err == nil {
						// 转换为百分比数值
						speedVal := attributes["attack_speed"].(float64) + (val / 100)
						attributes["attack_speed"] = speedVal
					}
				case "贪婪": // 增加秒杀几率
					if val, err := strconv.Atoi(cleanValue); err == nil {
						killVal := attributes["instant_kill"].(int) + val
						attributes["instant_kill"] = killVal
					}
				case "傲慢": // 增加最大HP
					if val, err := strconv.ParseFloat(cleanValue, 64); err == nil {
						// 转换为百分比数值，基于当前HP值增加
						hpVal := attributes["hp"].(int)
						attributes["hp"] = hpVal + int(float64(hpVal)*val/100)
					}
				case "嫉妒": // 增加暴击伤害
					if val, err := strconv.ParseFloat(cleanValue, 64); err == nil {
						// 转换为百分比数值
						damageVal := attributes["critical_damage"].(float64) + (val / 100)
						attributes["critical_damage"] = damageVal
					}
				case "色欲": // 自动回复
					if val, err := strconv.Atoi(cleanValue); err == nil {
						recoveryVal := attributes["recovery"].(int) + val
						attributes["recovery"] = recoveryVal
					}
				case "暴怒": // 增加暴击率
					if val, err := strconv.ParseFloat(cleanValue, 64); err == nil {
						// 转换为百分比数值
						rateVal := attributes["critical_rate"].(float64) + (val / 100)
						attributes["critical_rate"] = rateVal
					}
				case "怠惰": // 增加攻击力
					if val, err := strconv.ParseFloat(cleanValue, 64); err == nil {
						// 转换为百分比数值，基于当前攻击力增加
						attackVal := attributes["attack"].(int)
						attributes["attack"] = attackVal + int(float64(attackVal)*val/100)
					}
				}
			} else {
				// 处理普通附加属性
				switch attr.AttrType {
				case "hp":
					if val, err := strconv.Atoi(attr.AttrValue); err == nil {
						hpVal := attributes["hp"].(int) + val
						attributes["hp"] = hpVal
					}
				case "attack":
					if val, err := strconv.Atoi(attr.AttrValue); err == nil {
						attackVal := attributes["attack"].(int) + val
						attributes["attack"] = attackVal
					}
				case "attack_speed":
					if val, err := strconv.ParseFloat(attr.AttrValue, 64); err == nil {
						attackSpeedVal := attributes["attack_speed"].(float64) + val
						attributes["attack_speed"] = attackSpeedVal
					}
				case "move_speed":
					if val, err := strconv.Atoi(attr.AttrValue); err == nil {
						moveSpeedVal := attributes["move_speed"].(int) + val
						attributes["move_speed"] = moveSpeedVal
					}
				case "bullet_speed":
					if val, err := strconv.Atoi(attr.AttrValue); err == nil {
						bulletSpeedVal := attributes["bullet_speed"].(int) + val
						attributes["bullet_speed"] = bulletSpeedVal
					}
				case "drain":
					if val, err := strconv.Atoi(attr.AttrValue); err == nil {
						drainVal := attributes["drain"].(int) + val
						attributes["drain"] = drainVal
					}
				case "critical":
					if val, err := strconv.Atoi(attr.AttrValue); err == nil {
						criticalVal := attributes["critical"].(int) + val
						attributes["critical"] = criticalVal
					}
				case "dodge":
					if val, err := strconv.Atoi(attr.AttrValue); err == nil {
						dodgeVal := attributes["dodge"].(int) + val
						attributes["dodge"] = dodgeVal
					}
				case "instant_kill":
					if val, err := strconv.Atoi(attr.AttrValue); err == nil {
						instantKillVal := attributes["instant_kill"].(int) + val
						attributes["instant_kill"] = instantKillVal
					}
				case "recovery":
					if val, err := strconv.Atoi(attr.AttrValue); err == nil {
						recoveryVal := attributes["recovery"].(int) + val
						attributes["recovery"] = recoveryVal
					}
				case "trajectory":
					if val, err := strconv.Atoi(attr.AttrValue); err == nil {
						trajectoryVal := attributes["trajectory"].(int) + val
						attributes["trajectory"] = trajectoryVal
					}
				}
			}
		}
		fmt.Println()
	}

	// 计算皮肤属性
	fmt.Println("2. 计算皮肤属性:")
	fmt.Printf("   皮肤: %s\n", activeSkin.Skin.Name)
	fmt.Printf("   - 皮肤攻击: %d\n", activeSkin.Skin.Attack)
	fmt.Printf("   - 皮肤攻击速度: %d\n", activeSkin.Skin.AtkSpeed)
	fmt.Printf("   - 皮肤HP: %d\n", activeSkin.Skin.HP)

	attributes["hp"] = attributes["hp"].(int) + activeSkin.Skin.HP
	attributes["attack"] = attributes["attack"].(int) + activeSkin.Skin.Attack
	attributes["attack_speed"] = attributes["attack_speed"].(float64) + float64(activeSkin.Skin.AtkSpeed)
	attributes["critical_rate"] = attributes["critical_rate"].(float64) + activeSkin.Skin.CriticalRate
	attributes["critical_damage"] = attributes["critical_damage"].(float64) + activeSkin.Skin.CriticalDamage
	if activeSkin.Skin.AtkType > 0 {
		attributes["atk_type"] = activeSkin.Skin.AtkType
	}

	fmt.Println()

	// 显示计算结果
	fmt.Println("3. 最终属性计算结果:")
	fmt.Printf("   HP: %d\n", attributes["hp"].(int))
	fmt.Printf("   Attack: %d\n", attributes["attack"].(int))
	fmt.Printf("   AttackSpeed: %f\n", attributes["attack_speed"].(float64))
	fmt.Printf("   MoveSpeed: %d\n", attributes["move_speed"].(int))
	fmt.Printf("   BulletSpeed: %d\n", attributes["bullet_speed"].(int))
	fmt.Printf("   Drain: %d\n", attributes["drain"].(int))
	fmt.Printf("   Critical: %d\n", attributes["critical"].(int))
	fmt.Printf("   Dodge: %d\n", attributes["dodge"].(int))
	fmt.Printf("   InstantKill: %d\n", attributes["instant_kill"].(int))
	fmt.Printf("   Recovery: %d\n", attributes["recovery"].(int))
	fmt.Printf("   Trajectory: %d\n", attributes["trajectory"].(int))
	fmt.Printf("   CriticalRate: %f\n", attributes["critical_rate"].(float64))
	fmt.Printf("   CriticalDamage: %f\n", attributes["critical_damage"].(float64))
	fmt.Printf("   AtkType: %d\n", attributes["atk_type"].(int))
	if dr, exists := attributes["damage_reduction"]; exists {
		fmt.Printf("   DamageReduction: %s\n", dr.(string))
	}

	fmt.Println()
	fmt.Println("4. 与用户提供的示例数据对比:")
	fmt.Printf("   用户示例中的攻击: 305 vs 计算结果: %d\n", attributes["attack"].(int))
	fmt.Printf("   用户示例中的攻击速度: 11 vs 计算结果: %f\n", attributes["attack_speed"].(float64))
	fmt.Printf("   用户示例中的HP: 50 vs 计算结果: %d\n", attributes["hp"].(int))

	// 检查是否一致
	if attributes["attack"].(int) == 305 && attributes["attack_speed"].(float64) == 11.0 && attributes["hp"].(int) == 50 {
		fmt.Println("\n✅ 属性计算结果与用户示例数据一致！")
	} else {
		fmt.Println("\n❌ 属性计算结果与用户示例数据不一致！")
	}
}
