#!/bin/bash
# Script para mover controllers para internal/api/controllers/
# Parte da refatoração de estrutura do internal/

set -e

CONTROLLERS_DIR="internal/api/controllers"
API_DIR="internal/api"

echo "🔄 Iniciando migração de controllers..."
echo "📁 Destino: $CONTROLLERS_DIR"
echo ""

# Criar diretório se não existir
mkdir -p "$CONTROLLERS_DIR"

# Contador
moved_count=0
skipped_count=0

# Iterar sobre cada subdiretório em internal/api/
for dir in "$API_DIR"/*/; do
  entity=$(basename "$dir")

  # Ignorar diretórios especiais
  if [[ "$entity" == "controllers" || "$entity" == "invite" ]]; then
    echo "⏭️  Ignorando: $entity (especial)"
    ((skipped_count++))
    continue
  fi

  # Procurar *_controller.go
  controller_file="$dir${entity}_controller.go"

  if [ -f "$controller_file" ]; then
    echo "✅ Movendo: $entity_controller.go"
    mv "$controller_file" "$CONTROLLERS_DIR/"
    ((moved_count++))
  else
    echo "⚠️  Não encontrado: $controller_file"
    ((skipped_count++))
  fi
done

echo ""
echo "📊 Resumo:"
echo "   Movidos: $moved_count controllers"
echo "   Ignorados/Não encontrados: $skipped_count"
echo ""
echo "✅ Migração concluída!"
