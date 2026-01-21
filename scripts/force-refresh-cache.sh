#!/bin/bash
# Force clear cache and get fresh analysis with CacheExpiresAt field

echo "🔄 Forcing fresh analysis to populate CacheExpiresAt field..."
echo ""

# Get all PodSleuth resources
PODSLEUTHS=$(kubectl get podsleuths -o name 2>/dev/null)

if [ -z "$PODSLEUTHS" ]; then
    echo "❌ No PodSleuth resources found"
    exit 1
fi

echo "Found PodSleuths:"
echo "$PODSLEUTHS"
echo ""

# Add force-refresh annotation to each
for ps in $PODSLEUTHS; do
    echo "📝 Adding force-refresh annotation to $ps..."
    kubectl annotate $ps kubesleuth.io/force-refresh="$(date +%s)" --overwrite
done

echo ""
echo "✅ Force refresh annotations added!"
echo ""
echo "⏳ Waiting 5 seconds for operator to process..."
sleep 5

echo ""
echo "📊 Checking results..."
echo ""

# Check if CacheExpiresAt is now populated
for ps in $PODSLEUTHS; do
    NAME=$(echo $ps | cut -d'/' -f2)
    echo "Checking $NAME:"

    CACHE_EXPIRES=$(kubectl get $ps -o jsonpath='{.status.nonReadyPods[0].logAnalysis.cacheExpiresAt}' 2>/dev/null)

    if [ -z "$CACHE_EXPIRES" ] || [ "$CACHE_EXPIRES" == "null" ]; then
        echo "  ❌ CacheExpiresAt: NOT SET (still null)"
    else
        echo "  ✅ CacheExpiresAt: $CACHE_EXPIRES"
    fi

    ANALYZED_AT=$(kubectl get $ps -o jsonpath='{.status.nonReadyPods[0].logAnalysis.analyzedAt}' 2>/dev/null)
    echo "  📅 AnalyzedAt: $ANALYZED_AT"
    echo ""
done

echo "🎉 Done! Refresh your browser (F5) to see the updated cache expiration."
echo ""
echo "If CacheExpiresAt is still null, the operator needs to be redeployed with latest code:"
echo "  make docker-build docker-push IMG=your-registry/kubesleuth-operator:latest"
echo "  kubectl rollout restart -n kubesleuth-operator-system deployment/kubesleuth-operator-controller-manager"
