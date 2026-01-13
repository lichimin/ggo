package controllers

import (
	"errors"
	"ggo/models"
	"ggo/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MailController struct {
	db *gorm.DB
}

func NewMailController(db *gorm.DB) *MailController {
	return &MailController{db: db}
}

func (mc *MailController) GetMails(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	var mails []models.Mail
	if err := mc.db.Where("user_id = ?", userID.(uint)).Order("created_at desc").Find(&mails).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "获取邮件失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, mails)
}

func (mc *MailController) ClaimMail(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "无效的邮件ID")
		return
	}

	var claimedMail models.Mail
	var rewardResult gin.H

	txErr := mc.db.Transaction(func(tx *gorm.DB) error {
		var mail models.Mail
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&mail, uint(id)).Error; err != nil {
			return err
		}

		if mail.UserID != userID.(uint) {
			return errors.New("无权领取该邮件")
		}

		if mail.Status != 0 {
			return errors.New("邮件已领取")
		}

		switch mail.ItemType {
		case "", "none":
			rewardResult = gin.H{"type": "none"}
		case "gold":
			if mail.Num <= 0 {
				return errors.New("金币数量无效")
			}
			if err := tx.Model(&models.User{}).Where("id = ?", userID.(uint)).
				Update("gold", gorm.Expr("gold + ?", mail.Num)).Error; err != nil {
				return err
			}
			rewardResult = gin.H{"type": "gold", "num": mail.Num}
		case "diamond":
			if mail.Num <= 0 {
				return errors.New("钻石数量无效")
			}
			if err := tx.Model(&models.User{}).Where("id = ?", userID.(uint)).
				Update("diamond", gorm.Expr("diamond + ?", mail.Num)).Error; err != nil {
				return err
			}
			rewardResult = gin.H{"type": "diamond", "num": mail.Num}
		case "equipment":
			if mail.ItemID <= 0 {
				return errors.New("装备ID无效")
			}
			var tpl models.EquipmentTemplate
			if err := tx.First(&tpl, mail.ItemID).Error; err != nil {
				return errors.New("装备不存在")
			}
			userEquipment := models.UserEquipment{
				UserID:       userID.(uint),
				EquipmentID:  mail.ItemID,
				IsEquipped:   false,
				Position:     "backpack",
				EnhanceLevel: 0,
			}
			if err := tx.Create(&userEquipment).Error; err != nil {
				return err
			}
			if err := tx.Preload("EquipmentTemplate").Preload("AdditionalAttrs").First(&userEquipment, userEquipment.ID).Error; err != nil {
				return err
			}
			rewardResult = gin.H{"type": "equipment", "equipment": userEquipment}
		case "treasures":
			if mail.ItemID <= 0 {
				return errors.New("宝物ID无效")
			}
			if mail.Num <= 0 {
				return errors.New("宝物数量无效")
			}
			var treasure models.Treasure
			if err := tx.First(&treasure, mail.ItemID).Error; err != nil {
				return errors.New("宝物不存在")
			}

			var myItem models.MyItem
			result := tx.Where("user_id = ? AND item_type = ? AND item_id = ?", userID.(uint), "treasure", mail.ItemID).First(&myItem)
			if result.Error != nil {
				if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
					return result.Error
				}
				myItem = models.MyItem{
					UserID:    userID.(uint),
					ItemID:    mail.ItemID,
					ItemType:  "treasure",
					Position:  "backpack",
					Quantity:  mail.Num,
					IsActive:  true,
					SellPrice: 0,
				}
				if err := tx.Create(&myItem).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Model(&models.MyItem{}).Where("id = ?", myItem.ID).
					Update("quantity", gorm.Expr("quantity + ?", mail.Num)).Error; err != nil {
					return err
				}
				myItem.Quantity += mail.Num
			}

			rewardResult = gin.H{"type": "treasures", "treasure_id": mail.ItemID, "num": mail.Num}
		default:
			return errors.New("未知物品类型")
		}

		if err := tx.Model(&models.Mail{}).Where("id = ? AND status = 0", mail.ID).
			Update("status", 1).Error; err != nil {
			return err
		}

		mail.Status = 1
		claimedMail = mail
		return nil
	})

	if txErr != nil {
		if errors.Is(txErr, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "邮件不存在")
			return
		}
		utils.ErrorResponse(c, http.StatusBadRequest, txErr.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message": "领取成功",
		"mail":    claimedMail,
		"reward":  rewardResult,
	})
}

type SendMailRequest struct {
	UserIDs []uint `json:"user_ids" binding:"required,min=1"`
	Title   string `json:"title"`
	Content string `json:"content" binding:"required"`
	Type    string `json:"type" binding:"required"`
	ItemID  uint   `json:"item_id"`
	Num     int    `json:"num" binding:"required,min=1"`
}

func (mc *MailController) SendMail(c *gin.Context) {
	var req SendMailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	switch req.Type {
	case "gold", "diamond":
	case "equipment", "treasures":
		if req.ItemID == 0 {
			utils.ErrorResponse(c, http.StatusBadRequest, "item_id必填")
			return
		}
	default:
		utils.ErrorResponse(c, http.StatusBadRequest, "type无效")
		return
	}

	mails := make([]models.Mail, 0, len(req.UserIDs))
	for _, uid := range req.UserIDs {
		mails = append(mails, models.Mail{
			UserID:   uid,
			Title:    req.Title,
			Content:  req.Content,
			ItemType: req.Type,
			ItemID:   req.ItemID,
			Num:      req.Num,
			Status:   0,
		})
	}

	if err := mc.db.Create(&mails).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "发送失败: "+err.Error())
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "发送成功", "count": len(mails)})
}

func (mc *MailController) SendMailPage(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(mailSendHTML))
}

const mailSendHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <title>发送邮件</title>
  <style>
    body { font-family: Arial, Helvetica, sans-serif; margin: 20px; }
    .row { display: flex; gap: 12px; flex-wrap: wrap; }
    .col { flex: 1; min-width: 240px; }
    label { display:block; font-size: 12px; color: #333; margin: 10px 0 6px; }
    input, select, textarea, button { width: 100%; padding: 8px; box-sizing: border-box; }
    textarea { height: 110px; }
    button { cursor: pointer; }
    .hint { font-size: 12px; color: #666; margin-top: 6px; }
    pre { background: #f6f8fa; padding: 12px; overflow: auto; }
  </style>
</head>
<body>
  <h2>发送邮件</h2>

  <div class="row">
    <div class="col">
      <label>管理员Token（固定：1108）</label>
      <input id="token" value="1108" placeholder="1108"/>
      <div class="hint">点击“加载用户”会调用 /api/v1/admin/users</div>
      <button id="loadUsers">加载用户</button>
      <label>选择用户（可多选）</label>
      <select id="users" multiple size="10"></select>
      <div class="hint">也可在下方手动填写 user_ids（逗号分隔）</div>
      <label>user_ids（可选）</label>
      <input id="userIdsText" placeholder="例如：1,2,3"/>
    </div>

    <div class="col">
      <label>标题（可选）</label>
      <input id="title" placeholder="例如：系统补偿"/>

      <label>内容</label>
      <textarea id="content" placeholder="邮件内容"></textarea>

      <label>物品类型</label>
      <select id="type">
        <option value="gold">gold(金币)</option>
        <option value="diamond">diamond(钻石)</option>
        <option value="equipment">equipment(装备)</option>
        <option value="treasures">treasures(宝物)</option>
      </select>

      <div class="row">
        <div class="col">
          <label>item_id（装备模板ID/宝物ID）</label>
          <input id="itemId" type="number" min="0" value="0"/>
        </div>
        <div class="col">
          <label>num（数量）</label>
          <input id="num" type="number" min="1" value="1"/>
        </div>
      </div>

      <button id="send">发送</button>
    </div>
  </div>

  <h3>结果</h3>
  <pre id="result"></pre>

  <script>
    function authHeader() {
      const t = document.getElementById('token').value.trim();
      return t;
    }

    function setResult(obj) {
      document.getElementById('result').textContent = typeof obj === 'string' ? obj : JSON.stringify(obj, null, 2);
    }

    function selectedUserIds() {
      const select = document.getElementById('users');
      const ids = [];
      for (const opt of select.selectedOptions) {
        const id = parseInt(opt.value, 10);
        if (!Number.isNaN(id)) ids.push(id);
      }
      return ids;
    }

    function parseUserIdsText() {
      const raw = document.getElementById('userIdsText').value.trim();
      if (!raw) return [];
      return raw.split(',').map(s => parseInt(s.trim(), 10)).filter(n => !Number.isNaN(n) && n > 0);
    }

    document.getElementById('loadUsers').addEventListener('click', async () => {
      setResult('加载中...');
      try {
        const resp = await fetch('/api/v1/admin/users', {
          headers: { 'Authorization': authHeader() }
        });
        const data = await resp.json();
        if (!resp.ok || !data.success) {
          setResult(data);
          return;
        }
        const select = document.getElementById('users');
        select.innerHTML = '';
        for (const u of data.data || []) {
          const opt = document.createElement('option');
          opt.value = u.id;
          opt.textContent = (u.username || '-') + ' #' + u.id;
          select.appendChild(opt);
        }
        setResult({ message: '用户已加载', count: (data.data || []).length });
      } catch (e) {
        setResult(String(e));
      }
    });

    document.getElementById('send').addEventListener('click', async () => {
      setResult('发送中...');
      try {
        const userIDs = Array.from(new Set([...selectedUserIds(), ...parseUserIdsText()]));
        const payload = {
          user_ids: userIDs,
          title: document.getElementById('title').value,
          content: document.getElementById('content').value,
          type: document.getElementById('type').value,
          item_id: parseInt(document.getElementById('itemId').value, 10) || 0,
          num: parseInt(document.getElementById('num').value, 10) || 0
        };

        const resp = await fetch('/api/v1/admin/mails/send', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': authHeader()
          },
          body: JSON.stringify(payload)
        });
        const data = await resp.json();
        setResult(data);
      } catch (e) {
        setResult(String(e));
      }
    });
  </script>
</body>
</html>`
